# S9-D10 — Promote: claim fields + save

**Kind:** Claude Design board  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Promote flow — claim fields step  
**Implements later as:** PR **S9-12**  
**Depends on:** S9-D9 (shell); S9-10 (write with confidence + argument)  
**Related:** S9-D11 precedes this step on the join path; research-judgment model §4  
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

The last screen of each subject's step. The researcher confirms the claim and saves it:

```text
Step 2 of 2 · Claim
  Status       [ Accepted         ▾ ]
  Confidence   [ —                ▾ ]   (optional)
  Argument     ┌──────────────────────────────┐
               │ Same name and age as PER-…   │   (drafted from confirmed matches; editable)
               └──────────────────────────────┘
                                   [ Done ]  [ Save & next ]
```

**Next saves this step** (one transaction). Then the flow goes to the walk (S9-D12), or ends on **Done**.

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Status dropdown ships with one option | `Accepted` only this spike. Lay it out for **Provisional** and **Rejected**, which arrive later without a relayout. |
| Confidence is optional | Low / Moderate / High, or none. Orthogonal to status. |
| Argument | Free text; drafted from confirmed pairs on the join path; often empty on a new handle. |
| Save is per subject | A failure fails this step only; earlier steps stay saved. |
| Minting | New handle: this is the only step before save. |

### 2.1 What this board is not

- Not compare (S9-D11).
- Not editing saved claims.

---

## 3. Implementation gate (S9-12)

| Ships in **S9-12** | Does **not** ship there |
| --- | --- |
| Claim fields + Save / Done + failure state | Provisional / rejected options |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| CF-1 | Status is a Select with one option, visibly a dropdown (not a label). |
| CF-2 | Confidence Select with an explicit empty choice. |
| CF-3 | Argument TextArea with the drafted text marked as editable draft. |
| CF-4 | Primary action saves and continues. The board decides whether **Done** on this screen saves the step first or discards it; a discard asks through the leave guard. |
| CF-5 | Saving state, success (moves on), and failure (Callout, step kept) are designed. |
| CF-6 | Summary of what will be written: subject → handle (new or existing ref), pin count. |

---

## 5. Suggested frames

1. Mint path claim fields.
2. Join path with drafted argument and pin summary.
3. Status dropdown open (one option).
4. Saving.
5. Failure: subject already a member (race).

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Claim-fields step | Snowflake | **New** | `Features/Promote/PromoteClaimStep.swift` | |
| Select | Component | Ship | kit | Status, confidence. |
| Field + TextArea | Component | Ship | kit | Argument. |
| Callout | Component | Ship | kit | Failure. |
| Button | Component | Ship | kit | Save & next, Done. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| Hiding the status dropdown | The layout must reserve it now. |

---

## 7. Out of scope

- Compare
- Walk

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-12** against the board and inventory (kit first).
