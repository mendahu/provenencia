# S9-D11 — Promote: compare

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Promote flow — compare step (existing handle only)  
**Implements later as:** PR **S9-19**  
**Depends on:** S9-D9 (shell), S9-D10 (next step); S9-17 (comparison read + pins + backfill)  
**Related:** conclusion model §5.1 (confirmed matches, backfill), §5.3  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

This brief **extends S9-D9**'s frames. Reuse its layout; change only what this kind needs. The S9-D9 board is Claude Design *Promote flow* (`bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd`), `Promote flow.dc.html`; add these frames to it. It shipped as a workspace place in S9-11 ([`completed.md`](../completed.md#s9-d9--design-promote-shell--choose-target)).

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

When the chosen handle already has members, line up the incoming subject's Observations against each member's, Property by Property, and let the researcher confirm which support the match.

```text
Step 2 of 3 · Compare with PER-7KD45 · James Robins (2 records)

  name        this record: James Robins
              ☑ Birth record (CPR-7Z…): James Robins          compatible
              ☑ Marriage (CPR-3M…):    J. Robins              compatible
                                                  [ Accept all name ]
  age / birth this record: age 34 (1851)
              ☐ Birth record: 14 May 1817                     differs
                                           [ Skip comparison ]  [ Next ]
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Only on the join path | A new handle skips this step. |
| Compatible pairs start checked | The app's date / name rules decide compatibility; differing pairs start unchecked. |
| Checked = pinned | Both Observations are pinned on the new claim **and** on the member's existing claim (backfill). |
| Unchecked = not pinned | Not a refusal; nothing records *does not support*. Copy must not imply a rejection. |
| Several Observations per member | Each is its own row. |
| Cross-property suggestions | Census age vs birth date may be offered; confirming still pins the two Observations. |
| Skip is allowed | Accept with no pins. |

### 2.1 What this board is not

- Not reconciliation (choosing a value).
- Not editing Observations.

---

## 3. Implementation gate (S9-19)

| Ships in **S9-19** | Does **not** ship there |
| --- | --- |
| Compare step: rows, bulk accept per Property, skip | Reconciliation, Observation edits |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CM-1 | Grouped by Property; incoming value shown once per group, member rows beneath with their Source. |
| CM-2 | Checkbox per pair; compatible pre-checked; differing clearly marked. |
| CM-3 | Accept-all per Property; clear per pair. |
| CM-4 | Skip comparison → claim fields with zero pins. |
| CM-5 | Summary line: *N pairs will be pinned*. |
| CM-6 | Copy avoids *reject / does not support* language for unchecked pairs. |
| CM-7 | Scales to a handle with many members (scroll, grouping). |

---

## 5. Suggested frames

1. Two members, mostly compatible.
2. Differing date pair unchecked.
3. Member with several name Observations.
4. Cross-property suggestion (age vs birth).
5. Large handle (8+ members).
6. Skip.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Compare step | Snowflake | **New** | `Features/Promote/PromoteCompareStep.swift` | |
| Checkbox rows | Component | Ship | kit checkbox / Card rows | |
| Section header | Component | Ship | kit | Per Property. |
| Badge / Chip | Component | Ship | kit | compatible / differs. |
| Observation value display | Snowflake | Ship | `Features/EvidenceGraph/ObservationValueDisplay.swift` | Reuse value rendering. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A diff-style comparison component in kit | One call site. |

---

## 7. Out of scope

- Claim fields
- Walk

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-19** against the board and inventory (kit first).
