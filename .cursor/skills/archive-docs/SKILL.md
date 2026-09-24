---
name: archive-docs
description: >-
  Distills a closed Provenencia spike or shipped idea into a short decisions
  note and deletes operational archive (PR sequences, completed checklists,
  Claude Design briefs). Use when closing a spike, moving
  docs/deployment-plan/spike-N to archive/, parking a shipped docs/ideas/
  file, leftover re-home after a brainstorm, or when the user asks to
  distill, clean, or archive documentation.
---

# Archive docs (decisions, not git)

Closed spikes and shipped ideas keep **themes and decisions**. PR order,
checklists, and design briefs live in **git**. Do not duplicate them in
`docs/`.

Pattern example: [`docs/deployment-plan/archive/`](../../../docs/deployment-plan/archive/)
(Spikes 1–7). Policy one-liner: [`archive/README.md`](../../../docs/deployment-plan/archive/README.md).

Open-spike briefs still use [`add-design-brief`](../add-design-brief/SKILL.md).
Domain rules stay in [`docs/`](../../../docs/) (`interpretation-layer-data-model.md`,
etc.) — the archive **points** at those; it does not restated the schema.

**Docs-only: no `VERSION` bump.**

## When this applies

| Change | This skill |
| --- | --- |
| Close a spike (`spike-N/` → `archive/spike-N/`) | **Yes** |
| Distill an already-archived spike that still has briefs / `completed.md` | **Yes** |
| Move a shipped `docs/ideas/*.md` into `ideas/archive/` | **Yes** |
| Authoring an **open** Claude Design brief | No — `add-design-brief` |
| Implementing the last PR of a spike | No — finish the PR, then this skill |

Do **not** run this on an **open** spike (Spike 8 while it still has a plan).

## Spike close — checklist

```
- [ ] Spike is actually done (or explicitly stopped). Honesty in the README.
- [ ] Write/replace archive/spike-N/README.md as a **decisions** note (below)
- [ ] Keep a file only if it is still a **live contract** (not “how we shipped”)
- [ ] Delete: completed.md, deployment-plan.md, design/ (briefs + indexes),
      dogfood how-tos, PR-sequence diagrams, suggested-PR-title tables
- [ ] Distill shipped ideas that this spike finished (same keep/delete rules)
- [ ] Grep inbound links; retarget to the README or the live domain doc
- [ ] Slim docs/deployment-plan/README.md completed row (one line, no file list)
- [ ] Update docs/ideas/README.md if an idea moved
```

### Keep (live contract)

Keep a sibling file in the archive folder only when **skills or code still
implement against it** as behavior, not history. Spike 3 examples:

- [`navigation-history.md`](../../../docs/deployment-plan/archive/spike-3/navigation-history.md)
- [`omnibar-search.md`](../../../docs/deployment-plan/archive/spike-3/omnibar-search.md)

Strip that file’s “PR sequence / board handoff / incremental delivery” sections.
Leave stack rules, what counts as an entry, out-of-scope.

If the contract has already graduated to `docs/*.md` or a skill, **delete** the
idea/spike essay and point there.

### Delete (operational)

- `completed.md` step write-ups
- `deployment-plan.md` gates, ASCII PR trees, dogfood bars
- `design/` and `design/archive/` Claude Design prompts
- Per-PR “In / Out / Testable / Depends on” tables
- Numbered leftover lists once every item is re-homed (one leftover table on
  the distilled note is enough)

## Distilled spike README

Target: **~40–80 lines**. Match [`archive/spike-7/README.md`](../../../docs/deployment-plan/archive/spike-7/README.md).

```markdown
# Spike N — short name

**Done.** One sentence on what the spike proved. Live models: [links].

## Decisions

- Durable product/engine choice (not “S7-08 shipped the thin composer”).
- What we **refused** (descoped) if someone will ask again.
- Pointer to the next spike only if it owns a leftover (e.g. composer rethink → 8).
```

Do **not** list every PR id. Do not paste the dogfood bar. “Composer is Option B”
stays; “S7-D4 gated S7-08” goes.

## Shipped idea

When promoting `docs/ideas/foo.md` → `docs/ideas/archive/foo.md`:

1. Replace the body with **what shipped + decisions** (or a stub if the idea
   graduated to a domain doc — see `design-system-hardening.md`).
2. Drop problem essays, PR slices, and “display ideas we did not pick.”
3. Leftovers: one table → Spike / dogfood / another idea / descoped.
4. Update [`docs/ideas/README.md`](../../../docs/ideas/README.md) Current vs Archived.

Brainstorms that were a backlog (`interpretation-graph-ui`) become a **decisions
table + leftover map**, not a preserved 700-line essay.

## Retarget

Grep the old paths (`completed.md`, `design/archive/S*-D*`, `deployment-plan.md`
under that spike). Code comments may name old brief ids (`S7-D4`); that is fine
if they do not **link** to a deleted file. Skills and `docs/` links must work.

## After close

The open-spike folder is gone. `docs/deployment-plan/README.md` **Current**
points at the next spike only. Archive index is the completed table +
`archive/README.md` — not a list of deleted files.
