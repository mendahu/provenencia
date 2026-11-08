---
name: archive-docs
description: >-
  Distills a closed Provenencia spike into a short decisions note and deletes
  operational archive (PR sequences, completed checklists, Claude Design
  briefs). Use when closing a spike, moving docs/deployment-plan/spike-N to
  archive/, deleting or re-homing a shipped docs/ideas/ file, leftover re-home
  after a brainstorm, or when the user asks to distill, clean, or archive
  documentation.
---

# Archive docs (decisions, not git)

Closed spikes keep **themes and decisions** under `docs/deployment-plan/archive/`.
Shipped ideas are **deleted** from `docs/ideas/` (git and the spike / PR are the
record). Open leftovers move to another current idea, a spike note, a domain doc,
or a skill — there is no `ideas/archive/`. PR order, checklists, and design briefs
live in **git**. Do not duplicate them in `docs/`.

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
| Delete or re-home a shipped / abandoned `docs/ideas/*.md` | **Yes** |
| Authoring an **open** Claude Design brief | No — `add-design-brief` |
| Implementing the last PR of a spike | No — finish the PR, then this skill |

Do **not** run this on an **open** spike (Spike 8 while it still has a plan).

## Spike close — checklist

```
- [ ] Spike is actually done (or explicitly stopped). Honesty in the README.
- [ ] Write/replace archive/spike-N/README.md as a **decisions** note (below)
- [ ] Keep a file only if it is a live contract or the design reasoning a domain doc links ([Keep](#keep)); not “how we shipped”
- [ ] Delete: completed.md, deployment-plan.md, design/ (briefs + indexes), prs/ (PR plans), *.review.md (review-plan logs), *.ship.md (ship-pr logs),
      dogfood how-tos, PR-sequence diagrams, suggested-PR-title tables, app-health audits (once their items are triaged or the user drops them), shipped dogfood notes (`ux/`)
- [ ] Delete shipped ideas this spike finished; re-home any open leftovers (idea / spike / domain doc / skill)
- [ ] Grep inbound links; retarget each one (§ Retarget): to the archive README, a kept doc, or the live domain doc, or to plain text naming the PR
- [ ] Slim docs/deployment-plan/README.md completed row (one line, no file list)
- [ ] Update docs/ideas/README.md if an idea moved
- [ ] Every ship log's follow-ups triaged before the logs are deleted: an idea file, a row in the next spike's map, or dropped with a reason. A follow-up left only in a ship log disappears with it
- [ ] The close PR (S#-99) deletes its own plan, review log, and ship log in its last commit, the one that records the ship log. It writes no Landed note and no completed.md paragraph; the PR body carries any deviation
- [ ] The S#-99 plan carries the note's text and the Leftovers triage, so the implementer transcribes them
```

### Keep

Keep a sibling file in the archive folder in one of two cases, and delete it otherwise:

- **A live contract:** skills or code still implement against it as behavior, not history (Spike 3: [`navigation-history.md`](../../../docs/deployment-plan/archive/spike-3/navigation-history.md), [`omnibar-search.md`](../../../docs/deployment-plan/archive/spike-3/omnibar-search.md)). Strip its “PR sequence / board handoff / incremental delivery” sections; leave stack rules, what counts as an entry, out-of-scope.
- **The design reasoning behind a domain doc:** the rules live in named `docs/*.md`, a skill or the code, and the file keeps the why (Spike 11 [`comparison-layers.md`](../../../docs/deployment-plan/archive/spike-11/comparison-layers.md), Spike 12 [`source-structure.md`](../../../docs/deployment-plan/archive/spike-12/source-structure.md), Spike 13 [`evidence-graph-comparison.md`](../../../docs/deployment-plan/archive/spike-13/evidence-graph-comparison.md), Spike 14 [`summary-inputs.md`](../../../docs/deployment-plan/archive/spike-14/summary-inputs.md)). Its first lines name those docs and say they win where they differ, and the docs that teach the mechanism link it, or nothing leads a reader to it once the spike folder goes.

Either way the kept file cites no plan, test or decision id from the deleted plans. An essay whose rules have graduated and whose reasoning no domain doc needs is deleted, and its readers point at the domain doc.

### Delete (operational)

- `completed.md` step write-ups
- `deployment-plan.md` gates, ASCII PR trees, dogfood bars
- `design/` and `design/archive/` Claude Design prompts
- `prs/` PR plans (their Landed notes feed the decisions note first)
- Per-PR “In / Out / Testable / Depends on” tables
- Numbered leftover lists once every item is re-homed (one leftover table on
  the distilled note is enough)

## Distilled spike README

Target: **~30–80 lines**. Match the headings of [`archive/spike-12/README.md`](../../../docs/deployment-plan/archive/spike-12/README.md).

```markdown
# Spike N — short name

**Done** (closed YYYY-MM-DD). One sentence on what the spike proved. Live models: [links]. Design reasoning: [kept file], if any.

## Decisions

- Durable product/engine choice (not “S7-08 shipped the thin composer”).

## Refused

What was refused or descoped, if someone will ask again.

## Leftovers

| Item | Home |
| --- | --- |
| Open item | idea file, next spike's map row, or "Dropped: reason" |
```

Do **not** list every PR id. Do not paste the dogfood bar. “Composer is Option B”
stays; “S7-D4 gated S7-08” goes.

## Shipped or abandoned idea

When an idea in `docs/ideas/` is finished or dropped:

1. **Delete** the file. Do not keep an `ideas/archive/` copy.
2. Durable decisions that skills or code still need live in a **domain doc**, a
   **skill**, or the **code** — graduate them there before deleting if they are
   not already.
3. Open leftovers: one row or short note in another current idea, the next
   spike's map, or `structural-debt.md` — never only in the deleted file, and
   never in `docs/dogfood/`, which holds feedback from people using the app.
4. Update [`docs/ideas/README.md`](../../../docs/ideas/README.md) and retarget
   inbound links (spike READMEs, domain docs, skills).

## Retarget

Grep the whole spike folder from outside it, so plans, review logs, ship logs,
briefs and feature docs are all found, not only the map:
`rg -n 'spike-N/' --glob '!docs/deployment-plan/spike-N/**' .`
Code comments may name old brief ids (`S7-D4`); that is fine
if they do not **link** to a deleted file. Skills and `docs/` links must work.

A link to the spike's map or a feature doc that is deleted becomes a link to
the archive README's `#decisions` (or to the live domain doc that now holds the
rule). A link to a PR plan, brief, review log or ship log becomes plain text
that names it and its PR, `S13-03 ship log (in git with #666)`, with no link:
it has no archive copy, and a README link would land where the cited text is
not.
An app-health audit becomes plain text with its date and the commit it was
written at: `the second app-health audit (2026-10-28, at 3cc609c; deleted when
Spike 14 closed, in git history)`.

Find what a close breaks by comparing broken links on the base and on the branch: the base has broken links of its own, so only new ones count. Run from the repo root on the close branch:

```bash
base=$(git merge-base origin/main HEAD); tmp=$(cd "$(mktemp -d)" && pwd -P)
git archive "$base" | tar -x -C "$tmp"
python3 scripts/check-plan-docs.py --links $(cd "$tmp" && find . -name '*.md' | sed "s|^\./|$tmp/|") \
  | grep 'broken link' | sed -E "s|^$tmp/||; s|:[0-9]+: | |" | sort -u > "$tmp.base"
python3 scripts/check-plan-docs.py --links $(git ls-files '*.md') \
  | grep 'broken link' | sed -E "s|:[0-9]+: | |" | sort -u > "$tmp.branch"
comm -13 "$tmp.base" "$tmp.branch"
rm -rf "$tmp" "$tmp.base" "$tmp.branch"
```

It prints each new broken link as `file error: broken link: target`; line numbers are dropped so edits that shift lines do not show. Fix each by the two forms above.

## After close

The open-spike folder is gone. `docs/deployment-plan/README.md` **Current**
points at the next spike only. Archive index is the completed table +
`archive/README.md` — not a list of deleted files.
