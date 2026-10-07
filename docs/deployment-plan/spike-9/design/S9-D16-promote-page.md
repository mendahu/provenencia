# S9-D16 — Promote: one page (graph alignment)

**Kind:** Claude Design board (**rethink**)  
**Spike:** Provenencia Spike 9 (canonical entities MVP)  
**View:** Promote: the whole flow, as one page  
**Implements later as:** PR **S9-44**  
**Depends on:** S9-42 (proposal read), S9-43 (batch write)  
**Replaces:** S9-D9 (shell + choose target) and S9-D10 (claim fields), both built in S9-11 / S9-12; S9-D11 (compare) and S9-D12 (walk), never built  
**Related:** [`promote-graph-alignment.md`](../../../promote-graph-alignment.md) (the design); [`ideas/promote-matching.md`](../../../ideas/promote-matching.md) (why); conclusion model §5  
**Design system layers:** [`docs/design-system-layers.md`](../../../design-system-layers.md)  
**Skill:** [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md); [`add-ui-component`](../../../../.cursor/skills/add-ui-component/SKILL.md); [`add-workspace-place`](../../../../.cursor/skills/add-workspace-place/SKILL.md)

Paste this entire document into Claude Design as the requirements for one board/flow. Read shared product facts in [`README.md`](README.md) first.

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

### This brief is a RETHINK: throw out the existing Promote design

**The Promote flow is being replaced wholesale.** The board is Claude Design *Promote flow* (`bc84685e-bbc3-4053-a5c9-0f5ac7a13ccd`), `Promote flow.dc.html`.

1. **Delete every existing frame on the board** before drawing anything. That includes:
   - the shell and choose-target frames (S9-D9);
   - the claim-fields frames (S9-D10);
   - any compare frames (S9-D11) or walk frames (S9-D12), and any drafts of them.
2. **Don't carry over the old structure:**
   - no step row or step indicator;
   - no Next / Back footer;
   - no New / Existing radio pair as a first screen;
   - no per-Property checklist of Observation pairs;
   - no walk queue.

   If a frame you're drawing looks like a step of a wizard, it's the old design.
3. **Don't keep a before/after, and don't annotate what changed.** The old frames aren't a reference to improve on. The only things that survive are kit components (Button, Select, Badge, Card, Field…) and the Conclusion row header (`ConclusionListRow`), which are instanced fresh from the kit.
4. **The board's earlier decision that still holds:** Promote is a **workspace place**, not a sheet (it needs the page's width, and the graph stays behind it in history). Sheets are used *inside* this page, for a row's evidence.

---

### Shared Spike 9 facts (all Conclusion boards)

- A canonical **Person / Event / Place** (`PER-…` / `EVT-…` / `PLC-…`) is a researcher's handle for one historical thing. In the UI it is a Person, never a "canonical entity." Interpretation Subjects on Evidence graphs keep candidate refs (`CPR-…`).
- A handle's **members** are the Subjects promoted onto it (accepted Identity Claims). Each Property collects values from every member, and the engine **reconciles** them.
- An **Evidence graph** is one Source's Subjects (people, events, places) joined by **bridges** (participation, relationship, location), each carrying cited Observations.
- Promote only **creates** claims. Editing or removing claims is a later workflow (Spike 10).
- No likeness/photo value exists yet. Thumbnails are a slot with a per-kind placeholder.

---

## 1. Objective

Promote files an Evidence graph's Subjects onto Persons, Events and Places. Think of it as laying this Source's graph onto the tree: most Subjects line up with handles that already exist, a few branches are new information, and some aren't worth filing yet.

**The app proposes the whole graph alignment; the researcher fine-tunes it.** One page lists every person, event and place on the graph, each with the handle the app thinks it is, how sure it is and why. The researcher fixes the few rows that need judgment and presses **Done**. Everything is filed in one go.

```text
Promote · Obituary of Grace Gray Gates (Frickleton), Medicine Hat News
──────────────────────────────────────────────────────────────────────
  Gracie Gray Gates (Frickleton)  CPR-QPS5B   → [ PER-7KD45 Grace Gray Gates ▾ ]  Strong   ⋯
  Birth · 1 May 1901              CEV-8864X   → [ EVT-2M1 Birth of Grace Gates ▾ ]  Strong
  Death · 27 Dec 1990             CEV-TJ68V   → [ EVT-9Q4 Death of Grace Gates ▾ ]  Strong
  Bill Davies                     CPR-CN1RP   → [ PER-4HD2 Bill Davies ▾ ]          Weak  ⚠
  Marion Robins                   CPR-K4WA8   → [ New Person ▾ ]                     No match
  Bill and Aida's Home            CEV-AGMFP   → [ Skip ▾ ]                           No match
  Edmonton                        CPL-Y10SN   → [ PLC-EDM Edmonton ▾ ]               Strong
  …
  22 connections will be filed  ⌄                                  [ Cancel ]  [ Done ]
```

---

## 2. Domain facts

| Fact | UI implication |
| --- | --- |
| Entry is the Promote button on any unpromoted card | The page opens with **that Subject's row**, already matched to its best handle. |
| Promoting one Subject is the same page with one row | A clear action, **"Map the rest of this graph (N)"**, adds a row for every other person, event and place. No separate single and multi flows. |
| Already-promoted Subjects are anchors | Their rows show their existing handle **read-only**, with no controls. They're context, quieter than editable rows, but legible. |
| Each unpromoted row has a target | A dropdown: the best match **preselected**, the next few alternatives, **New {kind}**, **Skip** (leave it on the graph, unfiled). Weak and unmatched rows default to **Skip**. |
| Each row has an assessment | **Strong**, **Weak** or **No match**, with a short reason ("name, birth and death agree"; "via Gracie → PER-7KD45 (child of)"). |
| The evidence is inspectable | Clicking the assessment opens a **sheet**: every comparison the app made for this row (the Subject's own records and its neighbors' — a birth date reached through the birth event), each **agrees / conflicts / unknown**, with its weight. Agreeing ones are preselected as **pins** (the claim's evidence); the researcher can toggle any. |
| Each row is a claim | Status (one option, *Accepted*, this spike; laid out for more), confidence, argument (drafted from the agreeing comparisons; editable). These may live in the row's sheet or an expanding row. Keep the row itself light. |
| Rows depend on each other | Changing one row's target re-proposes the **suggested** rows. Rows the researcher has touched are never changed. Rows that moved show a brief **updated** mark with the reason. |
| Conflicts are flagged, not fixed | A decided row the new context contradicts keeps its choice and shows a warning. |
| Possible duplicates | Two rows on the same existing handle, or two near-identical New rows, get a warning ("these may be the same person; combine them on the Evidence graph"). |
| Bridges aren't rows | One summary line, "22 connections will be filed", which expands to a list with a switch per bridge (for a relationship the researcher doesn't accept). A bridge with a skipped end shows as unfiled, with why. |
| Done files everything at once | One write. If the catalog changed meanwhile, Done fails cleanly and the page re-proposes. |
| Leaving | The leave guard asks before discarding manual changes. Nothing else is saved; re-opening re-proposes. |

### 2.1 What this board is not

- Not the Evidence graph, its cards or the composer. Only the Promote entry point is touched there.
- Not editing or removing existing claims (Spike 10).
- Not a wizard, and not a checklist of Observation pairs.

---

## 3. Implementation gate (S9-44)

| Ships in **S9-44** | Does **not** ship there |
| --- | --- |
| The page, rows, target dropdown, assessment + evidence sheet with pins, claim fields, updated marks, warnings, bridge summary with switches, Done, leave guard | Provisional / rejected status; editing existing claims; learned weights; a saved draft |

---

## 4. Requirements

| ID | Requirement |
| --- | --- |
| PP-1 | One page, no steps. Header: the Source (title, ref), what is being promoted, and the row count. |
| PP-2 | Opens with the clicked Subject's row, matched. **Map the rest of this graph (N)** adds every other person, event and place. |
| PP-3 | Row: the Subject (kind mark, name or label, `CPR-…` ref) → target dropdown → assessment badge. Already-promoted rows read-only. |
| PP-4 | Target dropdown: best match preselected; alternatives as Conclusion row headers; **New {kind}**; **Skip**. |
| PP-5 | Assessment badge (strong / weak / no match) with a one-line reason; opens the evidence sheet. |
| PP-6 | Evidence sheet: comparisons grouped by the Subject itself and each neighbor reached ("Birth · date"); each shows both values, the Sources, agree / conflict / unknown and its weight; a pin toggle per comparison; the claim fields (status, confidence, argument with a drafted default). |
| PP-7 | Suggested vs decided: touched rows are marked as decided (subtle). Re-proposed rows show **updated** with the reason. |
| PP-8 | Warnings: conflict with a decided row; possible duplicate; self-link or refused bridge. |
| PP-9 | Bridge summary line, expandable to a list with switches; unfiled bridges say why. |
| PP-10 | Footer: Cancel (through the leave guard when there are manual changes) and **Done**. Done states what it will do ("File 12 · Skip 7 · 22 connections"). |
| PP-11 | Scales to the obituary (19 primary rows plus already-promoted anchors) and beyond: scrolling, grouping by kind or by assessment. |
| PP-12 | Copy never calls a skipped or unchecked item a rejection. Skip means "not filed yet". |

---

## 5. Suggested frames

1. **Single row:** a card's Promote just opened, matched Strong, with *Map the rest of this graph (18)*.
2. **The whole obituary mapped:** a mix of Strong, Weak, No match → Skip, New, and read-only anchors.
3. **Evidence sheet** for Gracie: her name, her birth's date and place through the birth event, her death; pins preselected.
4. **Evidence sheet with a conflict:** a child whose birth year disagrees, unpinned, with a warning.
5. **After a change:** the researcher retargets Bill; two suggested rows show **updated** with the reason, and a decided row shows a conflict warning.
6. **Bridge summary expanded:** switches, and one bridge unfiled because an end is skipped.
7. **Possible duplicate** warning.
8. **Done** in progress and done (toast, back to the graph); Done failing because the catalog changed, then re-proposed.
9. **Leave guard** with manual changes.

---

## 6. UI building-block inventory

This table is **binding**. Instance the Ship kit rows; do not redraw them. Paths from `macos/App/` unless noted.

| Building block | Layer | Status | Home | Notes |
| --- | --- | --- | --- | --- |
| Promote page | Snowflake | **Rethink** | `Features/Promote/PromoteView.swift` | Replaces the step shell; a workspace place. |
| Promote model | Snowflake | **Rethink** | `Features/Promote/PromoteModel.swift`, `PromoteFlow.swift` | A row list (suggested / decided), not a step machine. |
| Choose-target step, claim step | Snowflake | **Remove** | `Features/Promote/PromoteTargetStep.swift`, `PromoteClaimStep.swift` | Gone; their pieces move into the row and the sheet. |
| Graph alignment row | Snowflake | **New** | `Features/Promote/` | Subject → target → assessment. |
| Evidence sheet | Snowflake | **New** | `Features/Promote/` | A sheet over the page. |
| Conclusion row header | Snowflake | Ship (from D2) | `Features/Conclusions/ConclusionListRow.swift` | Candidates in the dropdown, anchors' handles. |
| Select / ComboBox | Component | Ship | kit | Target dropdown (ComboBox when alternatives need search). |
| Badge | Component | Ship | kit | Strong / Weak / No match; updated; decided. |
| Callout | Component | Ship | kit | Conflict, duplicate and refused-bridge warnings. |
| Field + TextArea / Select | Component | Ship | kit | Status, confidence, argument. |
| Card / SectionHeader | Component | Ship | kit | Sheet groups; page sections. |
| Disclosure | Component | Ship | kit | Bridge summary expand. |
| Button / Confirm / Toast | Component | Ship | kit | Done, Cancel, leave guard, after-save. |

### Explicit non-goals

| Do not add | Why |
| --- | --- |
| A wizard, stepper or step row | The flow is one page. |
| A kit "match row" or "evidence table" component | One call site; compose. |
| A per-Property pair checklist | Replaced by the evidence sheet's comparisons. |

---

## 7. Out of scope

- Editing, re-pinning or removing existing claims.
- Provisional / rejected status (laid out, not offered).
- Saving a draft between sessions.

---

## 8. Handoff

1. Archive this brief under `archive/` when the board is agreed.
2. Record in [`../completed.md`](../completed.md).
3. Implement **S9-44** against the board and inventory (kit first).
