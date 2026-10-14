---
name: ship-pr
description: >-
  Orchestrates one Provenencia PR from a Planned plan to a ready-for-review PR:
  dispatches the implementer, runs scripts/check-pr.sh itself, fans out five
  fresh reviewers (conformance and regression, maintainability, user
  experience, security, performance), dedupes and verifies their findings,
  writes one fix plan, and loops at most two review rounds before handing the
  PR to the user. Keeps the ship log. Use when the user asks to ship,
  implement, build, or run the implementation workflow for a planned PR
  (S#-NN), or to resume one.
---

# Ship a PR

This session is the **orchestrator**. It owns the loop, the ship log, and every
call that is not code. It does not write product code, and it does not review
code itself: the implementer builds, fresh reviewers judge, and this session
decides what happens to each finding.

```
orchestrator ──plan──▶ implementer ──report──▶ orchestrator ── runs scripts/check-pr.sh
     ▲                                              │
     │                                              ▼
     └── fix plan ◀── triage ◀── 5 reviewers in parallel (fresh context, read-only)
```

Skills the other roles load: [`implement-pr`](../implement-pr/SKILL.md),
[`review-pr`](../review-pr/SKILL.md). Plans come from [`plan-pr`](../plan-pr/SKILL.md).

## Roles and agents (Claude Code)

| Role | Agent | Model | Context |
| --- | --- | --- | --- |
| Orchestrator | This session | The session's model | The plan, the ship log, reports, findings |
| Implementer | `pr-implementer` (`.claude/agents/pr-implementer.md`) | Smaller | The plan or the round's fix plan. Nothing else |
| Reviewer | `pr-reviewer` (`.claude/agents/pr-reviewer.md`), once per lens | Strongest | The lens, the plan, base and head. Never the implementer's reasoning |
| Verifier | `pr-reviewer`, with a verify prompt | Strongest | One finding and the code |

Start each implementer and reviewer as a new agent. Do not continue an earlier
reviewer into round 2; a reviewer that saw round 1 defends it.

## Before you start

- The plan's **Status** is `Planned`, its **Review:** log exists, and its
  **Open questions** is `None.` Otherwise stop and tell the user; the plan goes
  back to [`plan-pr`](../plan-pr/SKILL.md).
- The branch in the plan header does not exist yet (or you are resuming: read
  the ship log and continue from its last entry).
- You are on the user's Mac: `scripts/check-pr.sh` needs `xcodebuild`. On
  another machine the gate cannot pass; stop and say so.

## Loop

```
- [ ] 0. Ship log: create prs/S#-NN-<slug>.ship.md from the template below
- [ ] 1. Implement: dispatch pr-implementer with the plan (round 0)
- [ ] 2. Gate: read the report; run scripts/check-pr.sh yourself
- [ ] 3. Review round 1: five pr-reviewers in parallel, full diff
- [ ] 4. Triage: dedupe, verify, decide each finding; write Fix plan 1
- [ ] 5. Implement Fix plan 1; gate
- [ ] 6. Review round 2: lenses with Fix-now findings, fix commits only
- [ ] 7. Triage round 2; Fix plan 2 only for blockers; implement; gate
- [ ] 8. Finish: PR body, ready for review, CI, report to the user
```

### 1. Implement

Dispatch `pr-implementer`:

```text
Implement the Provenencia PR planned in <plan path>. Follow
.agents/skills/implement-pr/SKILL.md, round 0. Return the report it defines.
```

If the report says **Stopped on**, stop the loop. Record it in the ship log and
take it to the user: a stop means the plan was wrong, and the plan is fixed
with `plan-pr`, not around it.

### 2. Gate

Run `scripts/check-pr.sh` yourself on the pushed head. Do not accept the
report's word for it.

- **PASS** → review.
- **FAIL** → send the failing check's output back to the implementer as a fix
  round of one item. Gate retries are not review rounds; after **three** failed
  gates in a row, stop and take the log to the user.

### 3. Review

Lenses: **all five on every PR** (the user's call, 2026-10-14; revisit for cost
with the routing table below).

| Lens | Rubric | Route by paths, if routing is turned on |
| --- | --- | --- |
| Conformance and regression | `review-pr/lenses/conformance.md` | Always |
| Maintainability | `review-pr/lenses/maintainability.md` | Always |
| User experience | `review-pr/lenses/ux.md` | `macos/App/**`, `Localizable.xcstrings`, `core/apperr`, new error codes |
| Security | `review-pr/lenses/security.md` | `core/ingest`, `core/objectpath`, `core/objectstore`, `core/derivatives`, `core/identity`, `core/catalogsession`, `api/ffi/**`, migrations, `*.entitlements`, URL or locator code |
| Performance | `review-pr/lenses/performance.md` | `core/database/**`, `core/search`, `core/derivatives`, graph, list, and cache code, `macos/App/Features/**` models |

Dispatch one `pr-reviewer` per lens, **in parallel**:

```text
Review a Provenencia PR through the <Lens> lens. Follow
.agents/skills/review-pr/SKILL.md and its rubric for this lens.
Plan: <plan path>. Base: <base sha>. Head: <head sha>. Round: 1.
Read-only. Return findings in the review-pr format, then the verdict.
```

Round 2 adds: `Round: 2. Ship log: <path>. Scope: <round-1 head>...<head>.`
Round 2 goes only to lenses that had a Fix-now finding.

### 4. Triage

Every finding gets exactly one decision, written in the ship log.

1. **Route.** Move each **Outside my lens** note to the lens it belongs to and
   treat it as a finding at the severity it deserves.
2. **Dedupe.** Same place and same problem from two lenses → one finding; keep
   the higher severity; list both IDs.
3. **Verify.** Read the code each finding names. Confirm it yourself when it is
   plain. For a **blocker** or **major** that is low confidence, surprising, or
   disputed by another lens, dispatch a verifier:

   ```text
   Try to disprove this finding against the code at <head sha>. Read-only.
   <finding>
   Answer: Confirmed / Refuted / Unclear, with the evidence.
   ```

   A conformance finding that carries a failing test is verified by the fix
   round itself (the implementer runs it red first); mark it Fix now.
4. **Decide.**

| Decision | When |
| --- | --- |
| **Fix now** | Verified blocker or major whose Fit is this PR. A minor only when the fix plan already touches that file and the change is a few lines |
| **Follow-up** | Fit is a preparatory PR or an idea; a Reshape; minors not fixed now. Goes to the PR body's follow-up list, and to `docs/ideas/` or the map if the user agrees |
| **Rejected** | Refuted, out of the plan's scope with no harm, or contradicted by a Decision or working rule. One line of evidence |
| **To user** | Contradicts a plan Decision, an AC, or a working rule but looks right; a product call; two lenses that disagree and the code cannot settle it. Pauses the loop |
| **Logged** | Nits |

A finding that contradicts the plan is never fixed by quietly changing the
plan. The plan's Decisions bind the orchestrator too.

5. **Write the fix plan** in the ship log, in the plan's Step shape:

   ```markdown
   **F-1 — One line** (from CON-2, MNT-4)
   Files: `path`. Skill: `add-…` if any.
   Failing test first: `TestX/case` — <code from the finding>
   Do: what to change, naming the existing code to reuse.
   Check: the test passes; scripts/check-pr.sh PASS.
   ```

   Order: blockers first, then majors, then the minors riding along. Keep it to
   what the findings need; this is not a second plan.

### 5–7. Fix rounds

Dispatch `pr-implementer` with `Ship log: <path>. Round: <n>. Follow
.agents/skills/implement-pr/SKILL.md, fix round.` Then the gate (step 2).

An `F-#` reported **Not reproduced** drops its finding as Rejected, with the
test output as evidence.

**Caps.** Two review rounds. Round 2 can only add **blockers**. After round 2,
the PR goes to the user with whatever is open; there is no round 3 without the
user. A Fix plan 2 exists only for round-2 blockers.

### 8. Finish

The loop is done when the gate passes and no **blocker** or **major** is open
(every one is fixed, deferred by an explicit Follow-up decision, or answered by
the user).

1. Update the PR body: what shipped against each user story, AC evidence,
   review summary by lens, follow-ups, deviations, the plan's **Manual check**,
   and links to the plan, its review log, and the ship log.
2. Commit the final ship log on the branch ("Record the S#-NN ship log.") and push.
3. `gh pr ready <n>`. CI runs on non-draft PRs. `gh pr checks <n> --watch`.
   A red CI check is a gate failure: one fix round, then back to the user if
   it stays red.
4. Report to the user (below). The user does the manual check, reviews, and merges.

## Ship log

Path: beside the plan, `prs/S#-NN-<slug>.ship.md`. Committed on the PR branch.
It is how the loop resumes, how the user audits it, and what the next planner
reads when something went wrong.

```markdown
# Ship log — S#-NN Title

**Plan:** [`S#-NN-<slug>.md`](S#-NN-<slug>.md) · **Branch:** `<branch>` · **PR:** #<n>
**State:** implementing | gate | review 1 | fixing 1 | review 2 | fixing 2 | ready | stopped — <why>

## Round 0 — implement
Head `<sha>`. Gate PASS. Report: <deviations, or "none">.

## Round 1 — review
| ID | Lens | Severity | Title | Decision | Evidence / reason |
| --- | --- | --- | --- | --- | --- |
| CON-1 | CON | blocker | … | Fix now → F-1 | verified: <how> |
| MNT-3 | MNT | major | … | Follow-up | preparatory PR; to user |
| SEC-2 | SEC | major | … | Rejected | `objectpath.Validate` already guards this, line 31 |

### Fix plan 1
**F-1 — …**

## Round 1 — fix
Head `<sha>`. Gate PASS. F-1 done; F-2 not reproduced (dropped CON-3).

## Round 2 — review
…

## Result
Gate PASS · CI <green> · open: none · follow-ups: MNT-3, UX-5 · to user: —
```

## Report to the user

```markdown
**S#-NN ready for review:** #<n>
- Shipped: one line per user story.
- Review: round count; per lens, findings fixed / follow-up / rejected.
- Needs you: To user items, each with your recommendation. "Nothing."
- Follow-ups: one line each, with a suggested home.
- Manual check: the plan's §10 steps.
```

If the loop stopped early, say where, why, and what you need.

## Never

- Write or edit product code. The implementer does; you write the ship log and the PR body.
- Let a reviewer see the implementer's report or reasoning.
- Fix a finding by changing the plan's Decisions, Scope, or ACs.
- Run a third review round, mark the PR ready with a blocker open, or merge.
