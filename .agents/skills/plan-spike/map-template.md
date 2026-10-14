# Spike templates

Two files for a new spike. Guidance in `<!-- -->` comments is for the planner;
delete it from the finished docs.

```text
docs/deployment-plan/spike-N/
  README.md             status, documents, not this spike
  deployment-plan.md    the PR map (below)
  <feature>.md          one design doc per idea or cluster
  completed.md          one short paragraph per landed PR, newest first
  design/README.md      brief schedule (add-design-brief)
  design/S#-D#-….md     briefs, written just before the PR they gate
  prs/S#-NN-<slug>.md   PR plans, written just before the PR (plan-pr)
```

---

## `README.md`

```markdown
# Spike N — Short name

## Status

**Open.** One sentence on where the spike stands. PR map: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md).

One or two sentences: what the last spike left, and what this one makes true.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | Goal, working rules, PR map, order |
| [**Feature name**](feature.md) | Design for … PRs are in the map |
| [**Completed**](completed.md) | Finished steps |
| [Design briefs](design/) | Claude Design, written just before the PR they gate |
| [PR plans](prs/) | One per PR, written just before the PR |

## Not this spike

What someone will ask about, and where it lives instead.
```

---

## `deployment-plan.md`

````markdown
# Deployment plan — Spike N

One paragraph: what this spike pulls in (dogfood notes, [ideas](../../ideas/…), leftovers) and the PR rule (one note, one PR, unless two edit the same part of one view or the same comparer).

**A brief covers one view.** A PR that changes chrome waits on a brief for each view it touches. Briefs and PR plans are written just before the PR ([`add-design-brief`](../../../.agents/skills/add-design-brief/SKILL.md), [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md)). A bugfix or a behavior change with no new chrome is not gated.

## Status

**Open.** Per-PR status is the [PR map](#pr-map). <!-- No per-PR status in prose. -->

> **Goal of this spike:** one sentence.

## Goal (dogfood bar)

**By spike close**, each of these is true on a real project:

1. **Outcome name.** What a person sees or can do. Concrete examples.
2. …

## Where the notes went

| Dogfood note or idea | PR | Brief |
| --- | --- | --- |
| … | S#-NN | S#-D# / — |

**Left in the dogfood log** — …

**Back in the ideas folder** — [`ideas/….md`](../../ideas/….md): ….

## Working rules

These are the calls for this spike. PR plans treat them as binding.

- **Rule in one sentence.** Reasoning and specifics: tables, types, keys, thresholds, examples. Binds S#-NN.
- **Refused thing is out.** S#-NN is descoped (date). Where it went.

## PR map

| PR | Title | Slice | Scope | Brief | Depends on | Plan | Status |
| --- | --- | --- | --- | --- | --- | --- | --- |
| S#-01 | Title | 1 | One line: what ships | — | — | — | Scoped |
| S#-02 | … | 1 | … | S#-D1 | S#-01, S#-D1 | [plan](prs/S#-02-slug.md) | Planned |
| S#-03 | … | 2 | … | — | — | [plan](prs/S#-03-slug.md) | Landed [#N](https://github.com/mendahu/provenencia/pull/N) |
| S#-04 | … | 2 | … | — | — | — | Descoped (date) → [idea](../../ideas/….md) |
| S#-99 | Dogfood close / docs | — | Bar pass, [`archive-docs`](../../../.agents/skills/archive-docs/SKILL.md) | — | all | — | Scoped |

<!-- Status words: Scoped, Blocked — on …, Planned, In progress, Landed #N, Descoped (date). See docs/deployment-plan/README.md. -->

## Design track

| Brief | View | Gates | Notes |
| --- | --- | --- | --- |
| S#-D1 | … | **S#-02** | … |

## Slices

**✎ = design brief**, just before the PR it gates. **Check** = what you can verify in the app when the slice lands.

```text
SLICE 1 — Name
  S#-01  Title
  ✎ S#-D1 ──▶ S#-02  Title
  Check: …

SLICE 2 — Name
  …

S#-99 — Dogfood close / docs
```

Which PRs can run in parallel: ….

### Dependencies at a glance

```text
S#-01 ──▶ S#-02
S#-D1 ──▶ S#-02
```

---

## PR scopes

Input to [`plan-pr`](../../../.agents/skills/plan-pr/SKILL.md). Two to six lines each. Link, do not restate.

### Slice 1 — Name

#### S#-01 — Title

**Ships.** One to three sentences of behavior.
**Not this PR.** … (S#-NN), … ([idea](../../ideas/….md)).
**Check.** What a reviewer sees.
**Context.** Working rule "…"; [feature doc §](feature.md#section); [domain doc §](../../model.md#section).
````

---

## `completed.md` entry (written by the implementing PR)

```markdown
## S#-NN — Title

**Landed YYYY-MM-DD** in [#N](https://github.com/mendahu/provenencia/pull/N). What shipped, in a few plain sentences a later planner can rely on.
```
