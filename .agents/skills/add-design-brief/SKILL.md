---
name: add-design-brief
description: >-
  Authors a Provenencia Claude Design brief (S*-D* under
  docs/deployment-plan/<open-spike>/design/, or docs/ideas/ when no spike is
  open) with the working-rules ritual, kit reach-for
  list, and a binding UI building-block inventory. Use when adding or changing
  a design brief, Claude Design board prompt, S8-D7-style paste doc, design
  gate, or when the user asks to scope UI in Claude Design before a PR.
---

# Add a Claude Design brief

Provenencia UI is designed in **Claude Design** before it is implemented. A brief
is the **entire prompt** — Claude Design will not follow a linked README unless
the text is in the pasted file. Copy the **Paste block** below verbatim,
immediately after “Paste this entire document…”.

Layers / implementing chrome: [`add-ui-component`](../add-ui-component/SKILL.md),
[`docs/design-system-layers.md`](../../../docs/design-system-layers.md).

## When this applies

| Change | This skill |
| --- | --- |
| New `S*-D*` brief / Claude Design board | **Yes** |
| Amending an open brief’s Claude instructions or inventory | **Yes** |
| Planning the PR the brief gates | No — [`plan-pr`](../plan-pr/SKILL.md), which calls this skill when the PR adds chrome |
| Implementing the PR the brief gates | No — the PR plan (`prs/`), `add-ui-component` (+ feature skills) |
| Evaluating an existing `PV*` | No — `evaluate-ui-component` |

## Checklist

```
- [ ] Path: docs/deployment-plan/<open-spike>/design/S#-D#-<slug>.md (or `docs/ideas/` when no spike is open)
- [ ] Header: Kind, Spike, Implements later as, Depends on, Related, layers, skills
- [ ] Paste line + **working-rules Paste block** (verbatim from this skill)
- [ ] In-place note: rethink (replace frames) vs enhancement (extend frames)
- [ ] Objective, domain facts, what this board is not
- [ ] Implementation gate table (what the PR ships)
- [ ] Numbered requirements + suggested frames
- [ ] **Wide layout** — each view's Regular and Wide (pane ≥ 1600pt) arrangement, a `-W` requirement, frames at both widths
- [ ] **UI building-block inventory** — situation → kit component (binding)
- [ ] Explicit non-goals: no new kit primitive for one call site
- [ ] Wired into spike design/README.md, deployment-plan.md (design track + checklist)
- [ ] python3 scripts/check-plan-docs.py <brief> reports no errors
- [ ] review-plan on the brief (inventory reuse, one interaction one kit piece); log beside it
- [ ] Docs-only: no VERSION bump
```

## Claude Design ritual (must be in the brief)

Claude Design often keeps a **stale design-system pack**. Every brief tells it to:

1. Clear the board’s local design-system cache
2. Delete the board’s design-system reference
3. Pull a **fresh** Provenencia design-system copy
4. Compose from that kit; bespoke only when the use is truly this domain

Work **in place**. Do not fork a parallel surface. Do not preserve an old design
when the brief is a rethink.

Copy the Paste block below — do not paraphrase it.

## Paste block

```markdown
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

### Two widths — draw both

The window opens at 1366 × 900pt but is often stretched across a large monitor.
Every view has two width classes, measured on the view's own pane (window minus
sidebar and any inspector), not the screen:

- **Regular** — pane below 1600pt. Draw at the default window: 1366 × 900, sidebar open.
- **Wide** — pane 1600pt and up. Draw at 2560 × 1440, sidebar open.

Rules for Wide:

1. **Use the extra space.** Show more at once: a second column of supporting content, the detail beside its list, more grid or table columns, an inspector docked open. Content jammed left with empty space beside it is never the Wide design. If the view truly has nothing to add, center its content group and say so on the frame.
2. **Keep readable things readable.** Prose, forms and the primary column keep their Regular maximum widths. Add columns; do not stretch text or fields.
3. **Lists, tables, grids and canvases fill the width.** Cards keep their size and add columns.
4. **Same content, same actions.** Width changes arrangement only. Anything in the Wide second column is stacked below, in an inspector, or behind a disclosure at Regular.
5. **Past every column's maximum, center the group** and let the gutters grow.

Draw the **Regular** and **Wide** frame of each view side by side, using the same kit pieces in both.
```

## Inventory (binding)

Every brief has a **UI building-block inventory**. It is how Claude knows what
exists and what to instance.

| Column | Write |
| --- | --- |
| Building block | Human name |
| Layer | Component / recipe / snowflake |
| Status | Ship / Extend / Rethink / Remove / New |
| Home | Swift or kit path |
| Notes | **When to reach for it** on this board |

Also fill **Reach for (kit)** in the Paste block (already there). The inventory
must **repeat the mapping for this surface** — e.g. “Auto Transcribe → Button on
Field, not a new Transcribe control”; “conflict mark → Badge”.

**New kit primitive** only if the inventory marks it New **and** a second call
site is already known. Otherwise snowflake + compose.

## Wide layout (binding)

Every brief says what each view does with a **Wide** pane (1600pt and up). The
breakpoint, the reasoning and the pattern table are in
[`docs/macos-client-patterns.md` § 9](../../../docs/macos-client-patterns.md#9-window-widths-regular-and-wide).
The Paste block carries the rules, because Claude Design will not follow the link.

In the brief, for **each view** the board touches:

| Write | Example |
| --- | --- |
| A **Wide layout** table: view → Regular → Wide → what moves where | “Source page — Regular: one column, Artifacts and Citations stacked — Wide: primary column + 400pt secondary column holding Metadata and Citations” |
| A requirement `XX-W1`… per view stating the Wide arrangement | “SP-W1: at Wide the Citations list sits in the secondary column, not below Artifacts” |
| A Regular and a Wide frame in **Suggested frames** | “Frame 2a Regular 1366 × 900 · Frame 2b Wide 2560 × 1440” |

Reach for the patterns in the doc's table first (list beside detail, primary +
secondary column, more grid or table columns, docked inspector, two-column form).
A view that centers instead of adding a column says why in one clause. An
enhancement brief on a view that has no Wide frame yet adds one.

Wide reuses the Regular frame's kit pieces. It is not a reason for a new kit
primitive. A layout helper the Wide frames share (a two-column page scaffold) is
a design-system question for [`add-ui-component`](../add-ui-component/SKILL.md),
raised once a second view needs it.

**A brief cannot waive a safety rule.** The board decides layout and wording.
It does not decide that a dirty form closes without asking, that a destructive
action skips its confirm, or that a failure says nothing. A frame that needs
one of those is an open question for the user, written into the brief. (S12-D3
let a link close a dirty graph form; the Spike 12 audit found the lost edits.)

## Brief shape

Match an open spike brief (rethink vs enhancement):

1. Title + metadata
2. Paste line + working-rules block (+ one sentence if rethink vs enhance)
3. Objective (what changes, what stays)
4. Domain facts table
5. Out / not-this-board
6. Implementation gate
7. Requirements (`XX-1`…, and `XX-W1`… for Wide)
8. Wide layout table
9. Suggested frames (Regular and Wide for each view)
10. Inventory + non-goals
11. Handoff (archive under `design/archive/`, `completed.md`, then the PR)

## After the board

- Archive the brief under the **open** spike’s `design/archive/`; write `completed.md`
- The PR plan (`prs/S#-NN-<slug>.md`, [`plan-pr`](../plan-pr/SKILL.md)) copies the requirement IDs and the inventory, then moves from `Blocked — on board` to `Planned`
- Implement against the board + inventory (kit first)
- Later briefs on the same surface **extend the new frames**, not the discarded ones

When a **spike closes**, use [`archive-docs`](../archive-docs/SKILL.md) — do not leave briefs and `completed.md` in `archive/`.
