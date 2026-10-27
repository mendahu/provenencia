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
         and may raise only new blockers. A fix made in round 1 is new plan
         text that round 1's reviewers never saw: round 2 checks it against
         the code like any other line
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

**Floors.** These keep a real finding from being answered away. That is the
most common reason an app-health audit later finds what a plan review already
saw: S12-13's review rated a Swift copy of an engine rule minor and logged it,
and the Spike 12 audit made it a top item.

- Rate with the lens's severity table, not the reviewer's word. A second
  mechanism, a domain rule in the wrong layer, a registry that is due, and a
  Decision that bends a working rule, a skill or a `.cursor/rules` line are
  **blockers**.
- A blocker of those kinds is answered **Fixed** or **To user**: never Logged,
  never Declined on the plan's own say-so, never moved to an Out row.

**A Decision that bends a rule is the user's.** When a Decision contradicts a
working rule, a skill or a `.cursor/rules` file, the writer does not settle it,
even when a reviewer's question led to it (S12-08 answered an executability
question by having FakeStore mirror the engine, against `swift-tests.mdc`). It
goes **To user**; if the user accepts, the plan's **Rules this plan bends**
names it and a Step updates the rule.

## Plan checks

The classes later audits found most often, each tied to a section of the
[`plan-pr` template](../plan-pr/template.md). The maintainability reviewer
works them on every PR plan, beside the rubric. A missing or empty section that
should not be empty is a finding at the severity the class carries.

| Check | Template | Ask |
| --- | --- | --- |
| Rules bent | §5 **Rules this plan bends** | Does a Decision, Out line or Step contradict a working rule, a skill or a `.cursor/rules` file? Each is listed there with the user's answer, or it is a blocker |
| Reach | §6 **Reach** | Does the PR add a way into existing code (a navigation path, an entry point) or make a state reachable (none, many, a deleted referent)? Is each destination checked: it resets on the new id, handles the missing referent, shows the state? |
| Scale | §6 **Scale** | For each write the PR adds or re-keys: what does it invalidate and reload at the map's **Scale**? A fan-out to every child of a parent, or one call per child, is a finding |
| Engine first | §3, §5 | Is a PR scoped to one layer carrying a multi-write user action, a refusal or a rule the other layer owns? The owning layer's change comes first |
| Replace, not beside | §6 **Deletes** | Does each new mechanism retire the one it replaces (the old key, field, list, preflight, patch path)? Is every callee whose last caller the PR removes in **Deletes**? |
| Sweep | §6 **Strings**, §9 | Does the DoD `rg` each renamed or deleted name, qualified, across skills, rules, docs and copy? When the PR changes what an action covers, does **Strings** list the copy that names the old scope? |
| Imports and locks | §7 | Does a Step move a type or call into another package (check the import graph), or take a lock, slot or session a caller may already hold? |

## Reviewer prompts

Fill in the angle brackets. Do not add the planning conversation.

**Maintainability**

```text
You are reviewing a Provenencia <PR plan | design brief | PR map> for long-term
maintainability, before anyone builds on it. Read-only: do not edit any file.

Document: <path>
Rubric: .agents/skills/review-pr/lenses/maintainability.md — read it first and
follow "How to work". On a PR plan, also work the "Plan checks" table in
.agents/skills/review-plan/SKILL.md, and the "Floors" there for severity.
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
| **blocker** | Adds debt that is expensive to undo later (wrong layer, a second mechanism beside an existing one even inside one package, bypassing a registry, a schema shape or per-type field the next named add must fight or repeat), a Decision or Out line that defers a known defect in code this PR adds to the later PR that will first reach it, or the implementer cannot proceed without a decision. Must be answered before Planned |
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
- **Scale.** Does the map name the sizes its work must stay fast at (its
  **Scale** line), and does each scope entry that adds a write or a list say
  its fan-out at that size?
- **Engine before client-only.** Is a PR scoped to one layer carrying a
  multi-write user action or a rule the other layer owns? Cut the owning
  layer's change first.
- **Another spike's code.** Does a PR change code that a PR in another open
  spike's map also changes? Name the dependency in both maps.
- **Docs and skills per PR.** Does each PR that renames, re-keys or retires a
  pattern name the docs and skills it makes false?

Findings are `K-#`, answered the same way, logged as `deployment-plan.review.md`.

## Learn from landings

When a PR lands, its plan's **Landed** note lists where the code left the plan.
When a deviation or a later rework traces back to a shape this review should
have caught, add one line to **Known traps** in [`maintainability.md`](../review-pr/lenses/maintainability.md)
or a question to [`executability.md`](executability.md). The rubric grows from
real misses, not from guesses.
