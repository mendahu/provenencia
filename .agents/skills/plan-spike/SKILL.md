---
name: plan-spike
description: >-
  Scopes a Provenencia spike or slice with the user and writes its PR map
  (docs/deployment-plan/spike-N/deployment-plan.md): goal and dogfood bar,
  feature context docs, spike-wide working rules, a table of PRs with one-line
  scopes, briefs, dependencies and status, slices with a Check each, and a
  short scope entry per PR. Does not write per-PR plans — those come later from
  plan-pr. Use when opening a spike, adding or re-cutting a slice, pulling
  ideas or dogfood notes into a spike, splitting a feature into PRs, re-planning
  after a PR lands or is descoped, or when the user asks for a roadmap,
  deployment plan, PR list, or PR map.
---

# Plan a spike or slice

This is the **design and scoping** half of the workflow. It starts wide (what
are we building, how does it fit the app, what is the architecture) and ends
with a **PR map**: every PR named, scoped in a line or two, ordered, and gated
by its briefs. The detailed plan for each PR is written later, one at a time,
just before that PR, with [`plan-pr`](../plan-pr/SKILL.md).

Why not plan every PR now: the first PRs change the code and the decisions
under the later ones. A map stays cheap to change; ten detailed plans go stale.

Layout and status words: [`docs/deployment-plan/README.md`](../../../docs/deployment-plan/README.md).
Map template: [`map-template.md`](map-template.md).

**Docs-only: no `VERSION` bump.**

## When this applies

| Change | This skill |
| --- | --- |
| Open a new spike | **Yes** |
| Add a batch to an odds-and-ends (even-numbered) spike | **Yes** — as a slice |
| Add a slice to the open spike, or re-cut one | **Yes** |
| Pull an idea from `docs/ideas/` or notes from `docs/dogfood/ux.md` into a spike | **Yes** |
| Descope, merge, or split PRs in the map | **Yes** |
| Write the detailed plan for one PR | No — [`plan-pr`](../plan-pr/SKILL.md) |
| Write a Claude Design brief | No — [`add-design-brief`](../add-design-brief/SKILL.md), usually from `plan-pr` |
| Close a spike | No — [`archive-docs`](../archive-docs/SKILL.md) |
| Park an idea that is not scheduled | No — a file in `docs/ideas/` |

## Odd and even spikes

Odd-numbered spikes (from 13) are large new features; even-numbered spikes are
odds and ends ([odd and even spikes](../../../docs/deployment-plan/README.md#odd-and-even-spikes)).
For an even spike, run the same phases per **batch**: the batch is a slice with
its own goal, bar outcomes (added under the spike's bar with the batch's name),
working rules prefixed with the batch (`(Batch 2) …`), and a feature doc when it
needs one. The spike's goal and its close (S#-99) stay as the template has them.

## How the session runs

This skill is a conversation, not a form. Expect several turns with the user.
Ask product questions; answer architecture questions from the code and the
domain docs, and say what you found. Write to files only when a section has
settled, so the user reviews text that is already agreed in principle.

Each phase below ends with something the user can react to. Do not jump to the
PR list before the goal and the working rules hold.

## Checklist

```
- [ ] 1. Frame: goal, dogfood bar, what is pulled in, what is not
- [ ] 2. Context: one feature doc per idea or cluster; read the domain docs and the code it touches
- [ ] 3. Working rules: every spike-wide decision written as a rule
- [ ] 4. Cut PRs: split rules below; IDs assigned; briefs identified
- [ ] 5. Sequence: slices with a Check each; dependency graph
- [ ] 6. Write the map from map-template.md
- [ ] 7. Review gate (below), then review-plan's map review
- [ ] 8. Wire: spike README, deployment-plan/README.md, design/README.md, ideas and dogfood sources
- [ ] 9. Report: the map, open product questions, which PR to plan first
```

## 1. Frame

Agree with the user, in a few sentences each:

- **Goal.** One sentence: what is true when the spike or slice is done.
- **Bar.** Numbered outcomes a person can check on a real project. These are the
  dogfood bar for a spike and the slice **Check** for a slice. Outcomes, not
  features: "José suggests Jose and the Person page still shows both spellings."
- **Pulled in.** Which ideas files, dogfood notes, and leftovers this covers.
- **Not this spike.** What someone will ask about, and where it lives instead.

## 2. Context

The PR plans draw on this, so it carries the design thinking.

- One **feature doc** per idea or cluster, in the spike folder. When an idea
  comes from `docs/ideas/`, move the file into the spike folder (the ideas
  README rule) and rewrite it as the design: problem, shape, data model, what
  stays later. Leftovers go back to an idea file.
- Feature docs explain the **shape**. They do not hold PR order or checklists;
  that is the map.
- Read the domain models in `docs/*.md` the feature touches and the code it
  changes. A map built without reading the code cuts PRs along the wrong seams.
  Note the packages, views, and tables each feature lands in.
- A domain rule that will outlive the spike belongs in the domain doc, not the
  feature doc. Note it; the PR that ships it updates the domain doc.

## 3. Working rules

Working rules are the spike's **decisions**: product calls, architecture,
stack choices, what was refused. `plan-pr` treats them as binding.

- One bolded sentence, then the reasoning and the specifics: names of tables,
  types, keys, thresholds, examples.
- Name the PR a rule binds when it binds one.
- Record a refusal as a rule ("**No canvas cover.** S10-48 is descoped …") so
  it is not reopened.
- Questions the user has not answered stay out of the rules. List them in the
  report. A PR that depends on one is `Scoped`, and `plan-pr` will ask it again.

## 4. Cut PRs

**One PR is one behavior a reviewer can check**, in the app or in a test. It
must also fit one plan a smaller model can carry out (about 10 steps).

Split rules:

- **Combine** two notes only when they edit the same part of one view or the
  same comparer. Otherwise one note, one PR.
- **Make the change easy first.** When the code a feature lands in needs
  reshaping (a registry that should exist, logic in the wrong layer, a copy
  about to become a third), cut a preparatory refactor PR ahead of the feature.
  It is reviewed and landed on its own, and the feature PR stays simple.
- **Engine before chrome.** A schema or core change the UI needs lands first,
  as its own PR, when the UI waits on a brief. A small engine change with no
  brief can ride with its UI.
- **A brief covers one view.** A PR that changes chrome waits on a brief for
  each view it touches. A bugfix or behavior change with no new chrome is not
  gated. Note the brief IDs (`S#-D#`); they are written later, just before the PR.
- **Split** a PR that changes more than one view plus its engine path, needs two
  migrations, or would have a migration and a large briefed UI change together.
- **Seeded vocabulary, migrations, and FFI** each have their own skill; a PR
  that needs several is a sign it is two PRs.
- **Descoping is a decision.** Keep the ID, mark it `Descoped (date)`, and say
  where the idea went.

IDs:

- PRs are `S{N}-NN`, numbered in the order they were cut, never reused. Gaps are
  fine. A later PR does not renumber earlier ones.
- Briefs are `S{N}-D{n}`, the same way.
- `S{N}-99` is spike close (dogfood pass, `archive-docs`).

## 5. Sequence

- **Slices** group PRs a reviewer checks together. Each slice has a **Check**:
  what you can see in the app when its last PR lands. Order slices so review
  stays one stream; say which PRs can run in parallel.
- **Dependencies** are real ones: code that must exist, or a board that must be
  agreed. Draw the graph in a `text` block. Ordering by preference is not a
  dependency.

## 6. Write the map

Path: `docs/deployment-plan/spike-N/deployment-plan.md`. A slice added to the
open spike edits that file. Copy [`map-template.md`](map-template.md) for a new
spike.

- The **PR map** table is the single place for per-PR status. Do not repeat
  status in prose; the Status section says only whether the spike is open and
  points at the table.
- The **Plan** column is `—` until `plan-pr` writes the plan.
- Each **scope entry** is short: Ships, Not this PR, Check, Context. It is the
  input to `plan-pr`, not the plan. Enough that a later session knows what the
  PR is for and which rules and doc sections to read. Two to six lines.
- Link, do not restate. A scope entry points at a working rule or a feature doc
  section instead of copying it.

## 7. Review gate

```
- [ ] Every bar outcome is made true by at least one PR, and every PR serves a bar outcome
- [ ] Every PR has a one-line scope, a Check, Depends on, and a brief or "—"
- [ ] Every PR that adds chrome names a brief for each view it touches
- [ ] No PR fails the split rules; the largest PR could be planned in about 10 steps
- [ ] Every Not-this-PR line and every refused idea has a home
- [ ] Working rules hold every decision a scope entry relies on; no open question is written as a rule
- [ ] Dependency graph and slice order agree
- [ ] Every pulled-in dogfood note or idea is in "Where the notes went"
```

Then run the **map review** in [`review-plan`](../review-plan/SKILL.md): a
fresh maintainability reviewer looks for preparatory refactors the map is
missing, two PRs building parallel mechanisms, a registry that is due, and a
pattern change with no skill update. Answer each finding in
`deployment-plan.review.md`. A preparatory PR it proposes is the user's call.

## 8. Wire

- Spike `README.md`: status, documents table, not this spike (template in the map template).
- [`docs/deployment-plan/README.md`](../../../docs/deployment-plan/README.md) **Current** points at the open spike.
- `design/README.md` **Scheduled** lists each brief ID, view, and the PR it gates.
- Remove pulled-in notes from `docs/dogfood/ux.md`; move pulled-in ideas and update [`docs/ideas/README.md`](../../../docs/ideas/README.md).

## 9. Report

Tell the user: the map path; the PRs by slice; the working rules you wrote that
they did not state themselves; open product questions; and which PR is ready
to plan first with `plan-pr` (its dependencies are met and its brief, if any, is
the next one to draw).

## Re-plan

The map changes as PRs land. When the user brings news (a PR landed, a board
moved the design, an idea was dropped):

- Update the row's status and the affected scope entries and working rules.
- Plans already written for affected PRs (`prs/`) get status `Blocked — re-plan`
  and are refreshed with `plan-pr`.
- Keep the IDs. Record a descope with its date and its home.
