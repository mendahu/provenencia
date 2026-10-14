---
name: review-plan
description: >-
  Reviews a Provenencia PR plan, Claude Design brief, or spike PR map before
  anyone builds on it, mainly for long-term maintainability: reuse of existing
  code, registries and patterns, the right layer, naming, near-copies,
  over-engineering, what the change makes dead, and whether the next named add
  gets a slot. Also dry-runs the plan as the implementing model. Runs
  scripts/check-plan-docs.py first, then fresh-context reviewers, then a
  bounded writer–reviewer loop with a review log. Use when a plan or brief is
  written and before it is marked Planned, when the user asks to review,
  critique, or second-opinion a plan, brief, or deployment plan, or from
  plan-pr, plan-spike, and add-design-brief.
---

# Review a plan

A plan that works is the minimum. This review asks whether it is the plan the
codebase **wants**: the one that reuses what exists, puts the change where the
concern lives, adds the fewest new concepts, deletes what it makes dead, and
leaves the next named change a slot instead of a fork. Implementing agents
drive straight at the instruction; the plan is the last cheap place to choose
the shape that ages well.

**Docs-only: no `VERSION` bump.**

## What gets reviewed

| Document | Lint | Lenses |
| --- | --- | --- |
| PR plan `prs/S#-NN-<slug>.md` ([`plan-pr`](../plan-pr/SKILL.md)) | Yes | Maintainability, Executability |
| Claude Design brief `design/S#-D#-<slug>.md` ([`add-design-brief`](../add-design-brief/SKILL.md)) | Yes | Maintainability (UI rows: Q1, Q3, Q5, Q6), Executability (Briefs) |
| PR map `deployment-plan.md` ([`plan-spike`](../plan-spike/SKILL.md)) | No | Maintainability, map questions below |

## Roles

| Role | Who | Context |
| --- | --- | --- |
| **Writer** | The session that wrote the plan | Owns the document and the review log. Runs the loop |
| **Maintainability reviewer** | The strongest model available. Read-only | Fresh: the document, the repo, [`maintainability.md`](../review-pr/lenses/maintainability.md). Nothing from the planning conversation |
| **Executability reviewer** | The model that will implement. Read-only | Fresh: the document, the repo, [`executability.md`](executability.md) |
| **User** | You | Reads the verdict, the review log summary, and anything sent **To user** |

Fresh context is the point: the implementer will not have the planning
conversation either, and a reviewer who watched the plan being written shares
its blind spots.

How to run a reviewer: in Claude Code, the Agent tool with a read-only agent
type and the model set for the role; in Cursor, a new agent chat or subagent
on the role's model. Give it the prompt below and nothing else.

## Loop

```
- [ ] 0. Lint: python3 scripts/check-plan-docs.py <doc>   → fix every error before any agent runs
- [ ] 1. Round 1: run both reviewers in parallel (prompts below)
- [ ] 2. Writer answers every finding in the review log: Fixed / Declined / To user
- [ ] 3. Writer applies the fixes; lint again
- [ ] 4. Round 2 (only if round 1 had a blocker or a Fixed reshape):
         fresh reviewers get the document + review log; they check the answers
         and may raise only new blockers
- [ ] 5. Stop: no open blocker, or round 2 is done. Anything still open → To user
- [ ] 6. Verdict in the review log and the plan's **Review:** line; report to the user
```

Two rounds, then the user. A reviewer can always find one more thing; the cap
keeps the loop from running until it agrees with itself.

### Answering a finding

| Answer | When | Write |
| --- | --- | --- |
| **Fixed** | The finding is right and fits this PR | What changed in the plan (section and ID) |
| **Declined** | The finding is wrong, or the cost is not worth it | The evidence: the path, the rule, or the map row that says otherwise. "I prefer my way" is not evidence |
| **To user** | A product call, a preparatory PR (map change), a Reshape, or a disagreement the writer cannot settle with evidence | The question and the writer's recommendation |

The writer does not give in to a finding it can refute, and does not refute one
it cannot. A product question is never settled between agents.

## Reviewer prompts

Fill in the angle brackets. Do not add the planning conversation.

**Maintainability**

```text
You are reviewing a Provenencia <PR plan | design brief | PR map> for long-term
maintainability, before anyone builds on it. Read-only: do not edit any file.

Document: <path>
Rubric: .agents/skills/review-pr/lenses/maintainability.md — read it first and
follow "How to work".
<Round 2 only: Review log: <path>. Check each answer to the round-1 findings.
Raise only new blockers.>

Assume the plan works. Your job is whether a better-fitting shape exists:
reuse, registries, the owning layer, names, copies, over-engineering, dead
code, the next named add. Search the repo before every claim; each finding must
name the existing code (path + symbol) or the map row / idea file it relies on.
No finding without evidence. Return the findings in the format of
.agents/skills/review-plan/SKILL.md "Finding format", most important first, then
a one-line verdict: Fits | Fits with changes | Reshape.
```

**Executability**

```text
You are the agent that will implement this Provenencia <PR plan | design brief>,
doing a dry run before you start. Read-only: do not edit any file, do not write code.

Document: <path>
Rubric: .agents/skills/review-plan/executability.md — read it first and follow it.
<Round 2 only: Review log: <path>. Check the answers; raise only new blockers.>

Report every place you would have to make a decision the plan did not make,
every path or symbol you could not find, every check you could not run, and
every Current state row that is not true. Use the format of
.agents/skills/review-plan/SKILL.md "Finding format", then a one-line verdict:
Ready | Ready with changes | Not ready.
```

## Finding format

```markdown
### M-3 · blocker · One-line title
**Plan:** STEP-4 / D-2 / Codebase fit row — what it says.
**Codebase:** `path` `Symbol` — what already exists or what the plan crosses. (Or: map row S#-NN / `docs/ideas/x.md`.)
**Better shape:** what to do instead.
**Why it ages better:** concretely — one place to change, N fewer copies, S#-NN becomes a registry entry, a concept fewer.
**Cost:** steps or lines added or saved; risk. **Fit:** this PR | preparatory PR | log it.
```

IDs: `M-#` maintainability, `X-#` executability, `K-#` map.

| Severity | Means |
| --- | --- |
| **blocker** | Adds debt that is expensive to undo later (wrong layer, a second mechanism beside an existing one, bypassing a registry, a schema shape the next named add must fight), or the implementer cannot proceed without a decision. Must be answered before Planned |
| **major** | A clear, cheap win in this PR's files. Fixed, or declined with evidence |
| **minor** | A judgment call. The writer may decline with one line |
| **nit** | Taste. Logged, not answered |

## Review log

Path: beside the document, `<doc-name>.review.md`. The implementer does not
read it; it keeps the plan short for the model that does. It holds the reasoning
the next planner and the user need.

```markdown
# Review — S#-NN Title

**Verdict:** Fits | Fits with changes | Reshape — Ready | Ready with changes | Not ready
**Rounds:** 1 | 2. **Reviewers:** maintainability (<model>), executability (<model>)

## To user
- Question, with the writer's recommendation. (Empty: "None.")

## Writer's own calls
- Decisions the writer made that no source dictated, so the user can veto them.

## Round 1
### M-1 · blocker · Title
<finding as returned>
**Answer:** Fixed — §6 Codebase fit row 2 now extends `X` instead of adding `Y`.
```

Then set the plan's **Review:** line to the log and the verdict. A plan goes to
`Planned` only when no blocker is open and **To user** is answered.
`check-plan-docs.py` refuses `Planned` without a linked log.

## Map review (plan-spike)

Run the maintainability reviewer on `deployment-plan.md` after the map's
review gate, with these questions in addition to the rubric:

- **Preparatory PRs.** Does a feature PR depend on code that should be
  reshaped first? Put the refactor in the map ahead of it, as its own PR.
- **Parallel mechanisms.** Do two PRs in this spike build the same kind of
  thing separately (two persistence paths, two list filters, two matchers)?
  One mechanism, two uses.
- **Registry due.** Do three or more PRs each add a hand-rolled case of the
  same thing? A registry PR comes first.
- **Pattern changes.** Does the spike change a pattern a skill teaches? A PR
  in the map updates the skill.

Findings are `K-#`, answered the same way, logged as `deployment-plan.review.md`.

## Learn from landings

When a PR lands, its plan's **Landed** note lists where the code left the plan.
When a deviation or a later rework traces back to a shape this review should
have caught, add one line to **Known traps** in [`maintainability.md`](../review-pr/lenses/maintainability.md)
or a question to [`executability.md`](executability.md). The rubric grows from
real misses, not from guesses.
