# Deployment plan

Working notes for implementation milestones. These are stopping points, not a full product roadmap.

Authoritative domain and stack decisions remain in the sibling docs under [`docs/`](../). Unscheduled product ideas live in [`ideas/`](../ideas/) — a parking lot, not a spike queue. UX friction from dogfooding lives in [`docs/dogfood/ux.md`](../dogfood/ux.md) until something is pulled into a spike, and then it leaves that log.

Closed spikes under [`archive/`](archive/) keep **themes and decisions**. PR order, checklists, Claude Design briefs, and PR plans are git history — not duplicated here.

## Odd and even spikes

From Spike 13, **odd-numbered spikes are large new features** and **even-numbered spikes are odds and ends**: small features, UX refinements and fixes. An odds-and-ends spike adds work in batches. Each batch is a slice in the spike's map with its own Check, bar outcomes and working rules (a rule names its batch), cut with [`plan-spike`](../../.agents/skills/plan-spike/SKILL.md) when it is pulled in. A one-off PR goes to the open even spike, as a new batch or into an open one. Work that needs its own dogfood bar or several briefs waits for the next odd spike.

## How a spike is planned

Two levels of plan, plus briefs and a review for each. The map is cut once and kept current. The plans are written one at a time, just before each PR, because earlier PRs move the code under later ones.

| Level | Skill | Output |
| --- | --- | --- |
| Spike or slice | [`plan-spike`](../../.agents/skills/plan-spike/SKILL.md) | `deployment-plan.md`: goal, working rules, **PR map**, slices, a short scope entry per PR. Feature design docs beside it |
| One PR | [`plan-pr`](../../.agents/skills/plan-pr/SKILL.md) | `prs/S#-NN-<slug>.md`: the whole task for the implementing agent |
| One view's chrome | [`add-design-brief`](../../.agents/skills/add-design-brief/SKILL.md) | `design/S#-D#-<slug>.md`, pasted into Claude Design. Gates the PR plan |
| Build one PR | [`ship-pr`](../../.agents/skills/ship-pr/SKILL.md) | The orchestrator runs [`implement-pr`](../../.agents/skills/implement-pr/SKILL.md), `scripts/check-pr.sh`, and five [`review-pr`](../../.agents/skills/review-pr/SKILL.md) lenses for at most two rounds; `prs/S#-NN-<slug>.ship.md` on the PR branch. Needs a `Planned` plan |
| Any plan above | [`review-plan`](../../.agents/skills/review-plan/SKILL.md) | `<doc>.review.md` beside it: a maintainability review and a dry run by fresh agents, at most two rounds. `python3 scripts/check-plan-docs.py` first |

```text
docs/deployment-plan/spike-N/
  README.md             status, documents, not this spike
  deployment-plan.md    the PR map
  <feature>.md          one design doc per idea or cluster
  completed.md          one short paragraph per landed PR
  design/               briefs (archive/ once the board is agreed)
  prs/                  PR plans
  *.review.md           review-plan logs, beside the doc they review
  prs/*.ship.md         ship-pr logs, committed on the PR branch
```

**PR status** is the same word in the map's Status column and the plan's header:

| Status | Means |
| --- | --- |
| `Scoped` | In the map with a scope entry. No plan yet |
| `Blocked — on …` | A dependency, a board, or a re-plan is outstanding |
| `Planned` | The plan passed the `plan-pr` gate and `review-plan`. An agent can implement it |
| `In progress` | `ship-pr` is running: a branch or draft PR is open |
| `Landed #N` | Merged. `completed.md` has the paragraph; the plan has a Landed note with any deviation |
| `Descoped (date)` | Not doing it. The map says where the idea went. The ID is not reused |

Spike 10 and earlier cut each PR as an In / Out / Testable table inside `deployment-plan.md`. A `plan-pr` plan for one of those PRs reads that table as its scope entry and adds `prs/` to that spike.

## Current

[Spike 11 — Promote matching refinement](spike-11/) is open. Plan: [`spike-11/deployment-plan.md`](spike-11/deployment-plan.md).

[Spike 12 — Source structure](spike-12/) is open alongside it. Plan: [`spike-12/deployment-plan.md`](spike-12/deployment-plan.md). Briefs are authored just before the PR they gate.

[Spike 13 — Person detail page](spike-13/) is open. Plan: [`spike-13/deployment-plan.md`](spike-13/deployment-plan.md). Each brief is designed, then implemented, before the next.

[Spike 14 — Odds and ends](spike-14/) is open. Plan: [`spike-14/deployment-plan.md`](spike-14/deployment-plan.md). Batch 1 is Event and Place pictures.

Ideas: [`ideas/`](../ideas/). Claude Design briefs: [`add-design-brief`](../../.agents/skills/add-design-brief/SKILL.md).

## Completed

| Spike | What we decided |
| --- | --- |
| [Spike 10](archive/spike-10/) | Dogfood fixes. Names match across accents and abbreviations. A face on a photo is a depiction value. |
| [Spike 9](archive/spike-9/) | Canonical entities via Identity Claims. One auto-reconciler cache. Promote is one page. |
| [Spike 8](archive/spike-8/) | Pause-and-refine data entry. Citation is the document. Named delete via Impact. |
| [Spike 7](archive/spike-7/) | Citation → Observation pipeline. Composer is a **place** (Option B). Subject types stay seeded. |
| [Spike 6](archive/spike-6/) | Spatial Evidence graph is **Go**. Neutral `GraphCanvas`. Connect was provisional until Spike 7. |
| [Spike 5](archive/spike-5/) | Interpretation data up to a graph stub. **No Subject UI.** One Sources family. |
| [Spike 4](archive/spike-4/) | Session cache + place registry. Go read-model tiering descoped. |
| [Spike 3](archive/spike-3/) | Persisted Back/Forward + omnibar (FTS5). Contracts: [`navigation-history.md`](archive/spike-3/navigation-history.md), [`omnibar-search.md`](archive/spike-3/omnibar-search.md). |
| [Spike 2](archive/spike-2/) | Source catalog + workspace chrome. Audit on first mutation. Files list descoped. |
| [Spike 1](archive/spike-1.md) | Local `*.provenencia` + install identity. cgo-SQLite-in-dylib risk retired. |
