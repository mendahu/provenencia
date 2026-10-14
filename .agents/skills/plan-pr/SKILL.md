---
name: plan-pr
description: >-
  Writes one Provenencia PR plan (docs/deployment-plan/<open-spike>/prs/S#-NN-<slug>.md)
  detailed enough that a less capable implementing agent can ship the PR
  without making a product or architecture decision: objective, user stories
  with acceptance criteria, scope, verified current state, binding decisions
  and stack choices, layer-by-layer design, ordered steps, named tests,
  definition of done, and stop conditions. Authors the gating Claude Design
  brief through add-design-brief when the PR adds chrome. Use when the user
  asks to plan, spec, scope, or write up a single PR, to turn a PR map row
  into an implementation plan, to refresh a stale PR plan, or to hand a PR
  to an agent.
---

# Plan a PR

A PR plan is the **whole task** for the agent that implements it. That agent
is usually a smaller model with no memory of the planning session. Everything
it needs to decide is decided here. Everything it needs to find is linked here
with a path. If the implementer would have to choose, the plan is not done.

The plan is written **just before the PR**, from the PR map row that
[`plan-spike`](../plan-spike/SKILL.md) left in `deployment-plan.md`. Not ten
plans up front: earlier PRs move the code and the decisions.

Layout and status words: [`docs/deployment-plan/README.md`](../../../docs/deployment-plan/README.md).
Template: [`template.md`](template.md). Copy it; keep every heading.

**Docs-only: no `VERSION` bump** for the plan itself.

## When this applies

| Change | This skill |
| --- | --- |
| Write the plan for a PR already in the map | **Yes** |
| Refresh a plan whose dependencies landed differently, or whose `main` moved | **Yes** |
| Plan a one-off PR outside a spike (bugfix, chore) | **Yes**. Put it under the open spike's `prs/` with the next free ID, and add a map row |
| Split a spike or slice into PRs, or change the PR list | No — [`plan-spike`](../plan-spike/SKILL.md) |
| Write the Claude Design brief the PR waits on | [`add-design-brief`](../add-design-brief/SKILL.md), called from step 3 below |
| Implement the PR | No. The plan's Implementer block is the implementer's instructions |

## Checklist

```
- [ ] 1. Locate: map row, scope entry, working rules, feature docs, slice Check
- [ ] 2. Ready? dependencies landed, map still says what the user wants
- [ ] 3. Brief: new chrome → agreed board exists, or author it now (add-design-brief)
- [ ] 4. Research the code: every touchpoint read, paths and symbols recorded, precedent found
- [ ] 5. Decide: every choice the implementer would face is a Decision; product calls asked
- [ ] 6. Write prs/S#-NN-<slug>.md from template.md, every heading kept
- [ ] 7. Self-review gate (below) passes, including the cold read
- [ ] 8. Wire: map row Plan link + status; scope entry matches; design/README if a brief
- [ ] 9. Report to the user: status, decisions you made, anything you changed in the map
```

## 1. Locate

Read, in this order:

1. The PR's row and scope entry in `docs/deployment-plan/<open-spike>/deployment-plan.md`.
   Spike 10 and earlier have no scope entries: read the PR's In / Out / Testable
   table instead, and treat Testable as the slice Check for that PR.
2. The map's **Working rules**. A rule that names this PR is binding on the plan.
3. The feature docs the scope entry links (the spike folder, then `docs/*.md`
   domain models). Read the sections, not just the titles.
4. The slice's **Check**. The plan's manual check must make it true.
5. `completed.md` and earlier plans in `prs/` for this view or package. Their
   **Landed** notes say where the code went somewhere other than the plan said.

## 2. Ready?

Stop and tell the user instead of planning when:

- A **Depends on** PR has not landed. Planning against code that does not exist
  yet produces a Current state that is fiction. Exception: the user asks for a
  draft. Then status is `Blocked — on S#-NN` and step 4 is redone on landing.
- The map row no longer matches what the user wants. Fix the map with
  `plan-spike` first; the plan and the map must agree.
- The PR is too big for one plan (see **Size** below). Propose the split.

## 3. Brief

The repo rule: a PR that adds or changes chrome waits on a Claude Design brief
for **each view** it touches. A bugfix or a behavior change with no new chrome
is not gated.

| Situation | Do |
| --- | --- |
| No new chrome | Brief field: `— (no new chrome)`. Say why in one clause |
| Brief agreed, board done | Link it. Copy its requirement IDs into **Scope** and its inventory into **Design → macOS** |
| Brief scheduled, not written | Write it now with [`add-design-brief`](../add-design-brief/SKILL.md). Plan status `Blocked — on board S#-D#` until the user says the board is agreed |
| Brief written, board not agreed | Plan status `Blocked — on board S#-D#`. Draft everything that does not depend on frames |

Ask the user how the agreed board reaches the implementer (exported frames in
the repo, screenshots, a link the implementer can open). Record it on the
**Board** line. If the implementer cannot see the board, **Design → macOS** must
describe each frame in words: layout, components from the inventory, states,
copy.

## 4. Research the code

This is most of the work. A plan built from the docs alone is a guess.

- Find **every file the PR will touch**. Search for the view, the store method,
  the FFI method, the table, the L10n namespace, the tests. Read them.
- Record **Current state** as `path` + symbol + what it does today + what is
  wrong with it for this PR. Paths are repo-relative. Symbols are exact.
- Find the **precedent**: the nearest earlier change of the same shape.
  `git log --oneline -- <path>`, `git log --grep=<word>`. Cite the commit or PR
  and say which part to copy ("seed it the way `engagement` was seeded in
  #382: `subjectvocab` `seedTerms`, no migration"). A precedent saves the
  implementer more than any paragraph.
- Load the **skills and rules** for the layers touched and note the rules that
  bind this PR. The implementer loads them too; name them in the header.

| Touches | Skill / rule |
| --- | --- |
| SQLite table or column | `add-catalog-migration`, `add-catalog-model`, `add-catalog-query` |
| Audited write, "last updated" | `add-audit-scope` |
| Delete of a resource | `add-catalog-delete` |
| Seeded terms, properties, types | `add-seeded-vocabulary` |
| Dates, names | `add-date-value`, `add-name-value` |
| Short refs (`SRC-…`) | `add-catalog-ref` |
| Omnibar hit kind | `add-searchable-kind` |
| FFI method | `add-ffi-handler`, `use-catalog-session` |
| Workspace place, list/detail, session cache | `add-workspace-place`, `add-workspace-location` |
| Any UI | `add-ui-component` (+ the brief's inventory) |
| User-facing copy | `add-localized-string` |
| Swift tests | `add-swift-test` |
| Go tests | `.cursor/rules/go-tests.mdc` (table-driven) |

Note the **planning base**: `git rev-parse --short HEAD` on an up-to-date
`main`. The implementer compares it with their checkout to spot drift.

## 5. Decide

List every point where an implementer would otherwise choose. Typical ones:
which package owns the logic, Go vs Swift, new table vs column vs derived
read, migration vs `Install`, new FFI method vs extending one, sync vs async,
session cache vs catalog, which kit component, error code, copy, test layer,
what happens on empty / failure / huge input.

- **Technical calls: make them.** Write each as a Decision with the reason and
  the alternative you rejected. Follow the domain docs, the working rules, and
  the precedent unless you write down why not.
- **Product calls: ask the user.** Batch the questions; give a recommendation
  for each. Do not write the plan around an open product question. A spike-wide
  answer also goes into the map's **Working rules**.
- **Open questions** in the template must be empty before status `Planned`.

A Decision is binding. The implementer may not reopen it. If it is wrong, they
stop and report (see the Implementer block). So decide only what the research
supports, and say in **Stop conditions** what evidence would prove it wrong.

## 6. Write

Path: `docs/deployment-plan/<open-spike>/prs/S#-NN-<slug>.md`. Slug matches the
suggested branch (`s10-41-sources-list`).

Copy [`template.md`](template.md). Keep every heading. When a section does not
apply, write `None —` and the reason, so the implementer knows it was
considered. The Implementer block is copied **verbatim**; do not paraphrase it.

Writing rules:

- **Plain declarative sentences**, the voice of the rest of `docs/`. "The list
  restores scroll after the sort is applied." Not "we should probably".
- **No soft words in a requirement.** `consider`, `maybe`, `ideally`, `e.g.`,
  `etc.`, `TBD`, `as appropriate`, `or similar`, `if needed` each hide a
  decision. Replace with the decision, or with a Stop condition.
- **IDs on everything that gets checked**: `US-1`, `AC-1.2`, `D-3`, `T-4`,
  `STEP-5`. The implementer reports against them, and the reviewer can cite them.
- **Concrete examples beat rules.** Each acceptance criterion has an input and an
  expected result: `Smith, John Henry` → Surname, Given, Given.
- **Exact names.** File paths, types, functions, FFI methods, L10n keys, error
  codes, test names. Mark new ones *(new)*.
- **Short code sketches are allowed** in Design for signatures, DDL, and proto
  messages. Do not write the implementation in the plan.
- **Out has a home.** Every Out line says where it goes: a later PR id, an idea
  file, the dogfood log, or "not planned".

### Size

One PR is one behavior a reviewer can check. A plan has reached the limit when:

- Steps go past about **10**, or
- it changes **more than one view** plus that view's engine path, or
- it needs **two migrations**, or a migration plus a large UI change that has its own brief.

Then propose a split to the user and update the map with `plan-spike`. A big
plan for a small model fails in the middle; two plans each land.

## 7. Self-review gate

Do not mark a plan `Planned` until each line is true. Fix the plan, not the gate.

```
- [ ] Every path in Current state, Design, Steps, and Tests exists, or is marked (new)
      → check with ls / rg; do not trust memory
- [ ] Every symbol named exists at the planning base, or is marked (new)
- [ ] Every In line is covered by at least one Step
- [ ] Every AC has a test (T-#) or a manual check line; every test cites an AC or a Decision
- [ ] Every Out line has a home
- [ ] Every Step names files, the skill to load if any, and a check command or observable result
- [ ] Commands are the repo's real ones (README, skills): go test -tags fts5, xcodebuild test, generate-proto.sh
- [ ] No soft words (list above) in Scope, Decisions, Design, Steps, or DoD
- [ ] Version bump line is filled in and matches docs/versioning.md
- [ ] Docs updates the PR owes are in the DoD (completed.md, map status, brief archive, domain doc)
- [ ] Stop conditions name concrete evidence, not "if something seems off"
- [ ] Open questions is empty, or status is Draft / Blocked
- [ ] Plan agrees with the map row and the working rules; any change was made in the map too
```

**Cold read.** Reread the plan as the implementer: you have this file, the
linked files, and the repo, and nothing from this conversation. At each step,
ask "what would I have to choose here?" Each answer is a missing Decision or a
missing path. Then ask "what would I be tempted to add?" Each answer is an Out line.

## 8. Wire

- Map row: **Plan** column links the file; **Status** is `Planned` or `Blocked — …`.
- If planning changed scope, update the map's scope entry in the same commit
  and tell the user what moved.
- A brief written here goes in `design/README.md` **Scheduled**, per `add-design-brief`.
- A spike-wide product answer goes in the map's **Working rules**.

## 9. Report

Tell the user, briefly: the plan path and status, the Decisions you made on your
own (so they can veto them before an agent builds on them), any map change,
and what blocks it if it is blocked.

## Refresh an existing plan

When a dependency lands differently or `main` has moved a lot:

1. Read the dependency's **Landed** note and `completed.md`.
2. Redo step 4 for every Current state row. Update the planning base.
3. Recheck each Decision against the new code. A changed Decision is a change
   to tell the user about.
4. Run the self-review gate again.

## After the PR lands

The implementer's DoD does this; check it was done:

- Plan status `Landed in #N (YYYY-MM-DD)`. Add a short **Landed** note at the top
  that records only where the code differs from the plan. Later plans read it.
- Map row status, `completed.md` entry, brief moved to `design/archive/`.

At spike close, [`archive-docs`](../archive-docs/SKILL.md) deletes `prs/` with
the rest of the operational docs. Decisions worth keeping go in the distilled README.
