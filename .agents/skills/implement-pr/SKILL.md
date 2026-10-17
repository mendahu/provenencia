---
name: implement-pr
description: >-
  Implements one Provenencia PR from its plan (prs/S#-NN-<slug>.md) or a fix
  round from its ship log: branch, follow the Steps, commit per behavior, pass
  scripts/check-pr.sh, open or update the draft PR, and report Definition of
  done evidence to the orchestrator. Deterministic checks only; no
  self-review. Use when ship-pr dispatches the implementer, or when the user
  asks to implement a planned PR or apply a fix plan.
---

# Implement a PR

You build what the plan says, prove it with the checks, and report. You do not
review your own work; fresh reviewers do that ([`review-pr`](../review-pr/SKILL.md)).
You do not decide product or architecture; the plan did.

## Inputs

- **Round 0:** the plan, `docs/deployment-plan/<spike>/prs/S#-NN-<slug>.md`.
- **Fix round:** the ship log, `S#-NN-<slug>.ship.md`, and the round number. Your
  task is that round's **Fix plan** (`F-#` items). The plan's Decisions and Out
  lines still bind.

## Round 0

1. **Read the plan's Implementer block** and obey it. It is the contract: the
   plan is the whole task, Decisions are made, do only In, stop on a stop condition.
2. **Branch.** `git fetch origin main`, then `git switch -c <Branch from the plan header> origin/main`.
   Compare `main` with the plan's **Planned** commit; if a **Current state** row is
   no longer true, stop and report which.
3. **Load every skill** in **Skills to load**.
4. **Work the Steps in order.** Run each step's Check before the next. A failing
   Check is fixed in that step.
5. **Commit per behavior.** The subject is one full imperative sentence ending
   in a period that says what changed for the researcher, like the rest of
   `git log`. No WIP commits. Never amend or force-push once pushed.
6. **Run the gate:** `scripts/check-pr.sh`. Fix until it prints `check-pr: PASS`.
   Never skip, disable, or weaken a test or a check to get there.
7. **Push and open a draft PR:**
   `git push -u origin <branch>` then
   `gh pr create --draft --base main --title "<plan §13 title>" --body-file <file>`.
   Body: plan §13 — objective, each user story with its AC evidence, deviations,
   manual check, and `Plan: <plan path>`.
8. **Docs lines of the DoD**, now that the PR has a number, the way the repo
   already records a landing before merge: the `completed.md` paragraph, the map
   row and the plan **Status** set to `Landed #N`, the plan's **Landed** note
   (only where the code differs from the plan), the brief moved to
   `design/archive/` if gated. One commit, "Record S#-NN as landed in pull
   request N.", then the gate again, then push. A spike close (S#-99) skips this step: its plan and completed.md are deleted by the PR ([`archive-docs`](../archive-docs/SKILL.md)).
9. **Report** (below).

## Fix round

1. Pull the branch. Read the ship log: the round's **Fix plan** and the findings
   each item cites.
2. Work the `F-#` items in order. For an item that carries a **failing test**:
   add the test first, run it, and confirm it **fails**. If it passes before
   any fix, do not change code for that item; report `F-# not reproduced` with
   the output. Then fix, and confirm it passes.
3. New commits only, one per item or per behavior. Never amend or force-push.
4. If a fix changed what shipped, update the `completed.md` paragraph and the
   plan's **Landed** note in the same round.
5. `scripts/check-pr.sh` until PASS. Push.
6. Report.

## Never

- Review or second-guess your own diff for other lenses. That is the reviewers' job.
- Implement an Out line, refactor code no Step or `F-#` names, or "tidy while here".
- Reopen a Decision. If one cannot work, stop and report the ID and the evidence.
- Skip, disable, mark flaky, or delete a test. Change a check to pass.
- Edit the plan's Decisions, Scope, or ACs. Report instead.
- Comment on GitHub, mark the PR ready, or merge. The orchestrator and the user do that.

## Report

Return this to the orchestrator, and nothing longer:

```markdown
## Implementer report — S#-NN, round <n>

**Branch / PR:** `<branch>` · #<n> (draft) · head `<sha>`
**Gate:** check-pr: PASS (paste the PASS/FAIL lines)

| Item | Status | Evidence |
| --- | --- | --- |
| AC-1.1 / F-2 | Done | `TestX/case` passes; commit `<sha>` |
| F-3 | Not reproduced | failing test passed before the fix: <output> |

**Deviations:** where the code differs from the plan or fix plan, and why. "None."
**Stopped on:** the stop condition or Decision and the evidence, or "—".
```
