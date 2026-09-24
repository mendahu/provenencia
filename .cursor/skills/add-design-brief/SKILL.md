---
name: add-design-brief
description: >-
  Authors a Provenencia Claude Design brief (S*-D* under
  docs/deployment-plan/…/design/) with the working-rules ritual, kit reach-for
  list, and a binding UI building-block inventory. Use when adding or changing
  a design brief, Claude Design board prompt, S8-D7-style paste doc, design
  gate, or when the user asks to scope UI in Claude Design before a PR.
---

# Add a Claude Design brief

Provenencia UI is designed in **Claude Design** before it is implemented. A brief
is the **entire prompt** — Claude Design will not follow a linked README unless
the text is in the pasted file.

Working-rules source (copy the Paste block verbatim):
[`docs/deployment-plan/spike-8/design/claude-design-working-rules.md`](../../../docs/deployment-plan/spike-8/design/claude-design-working-rules.md).

Layers / implementing chrome: [`add-ui-component`](../add-ui-component/SKILL.md),
[`docs/design-system-layers.md`](../../../docs/design-system-layers.md).

## When this applies

| Change | This skill |
| --- | --- |
| New `S*-D*` brief / Claude Design board | **Yes** |
| Amending an open brief’s Claude instructions or inventory | **Yes** |
| Implementing the PR the brief gates | No — `add-ui-component` (+ feature skills) |
| Evaluating an existing `PV*` | No — `evaluate-ui-component` |

## Checklist

```
- [ ] Path: docs/deployment-plan/<spike>/design/S#-D#-<slug>.md
- [ ] Header: Kind, Spike, Implements later as, Depends on, Related, layers, skills
- [ ] Paste line + **working-rules Paste block** (verbatim from claude-design-working-rules.md)
- [ ] In-place note: rethink (replace frames) vs enhancement (extend frames)
- [ ] Objective, domain facts, what this board is not
- [ ] Implementation gate table (what the PR ships)
- [ ] Numbered requirements + suggested frames
- [ ] **UI building-block inventory** — situation → kit component (binding)
- [ ] Explicit non-goals: no new kit primitive for one call site
- [ ] Wired into spike design/README.md, deployment-plan.md (design track + checklist)
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

Copy the Paste block from `claude-design-working-rules.md` — do not paraphrase it.

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

## Brief shape

Match an open Spike 8 brief (`S8-D7` rethink, `S8-D1` enhancement):

1. Title + metadata
2. Paste line + working-rules block (+ one sentence if rethink vs enhance)
3. Objective (what changes, what stays)
4. Domain facts table
5. Out / not-this-board
6. Implementation gate
7. Requirements (`XX-1`…)
8. Suggested frames
9. Inventory + non-goals
10. Handoff (archive under `design/archive/`, `completed.md`, then the PR)

## After the board

- Archive the brief under the **open** spike’s `design/archive/`; write `completed.md`
- Implement against the board + inventory (kit first)
- Later briefs on the same surface **extend the new frames**, not the discarded ones

When a **spike closes**, distill that folder to a decisions README. Delete `completed.md`, `deployment-plan.md`, and design briefs. Git has the PR history. See [`docs/deployment-plan/archive/`](../../../docs/deployment-plan/archive/).
