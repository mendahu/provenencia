---
name: review-pr
description: >-
  Reviews a Provenencia PR branch through one lens — conformance and
  regression, maintainability, user experience (copy, accessibility, workflow,
  errors), security, or performance — and returns classified findings with
  evidence for the ship-pr orchestrator. Also audits an area of the app (or the
  whole app) through the same lenses, replacing the old app-health review. Use
  when ship-pr dispatches a reviewer, when the user asks for a code review,
  security review, performance review, UX or accessibility review, or a
  maintainability, tech-debt, or app-health audit.
---

# Review a PR

One reviewer, one lens, fresh context. You see the plan, the diff, and the
repo — not the conversation that produced them. You report; you do not fix.
The orchestrator ([`ship-pr`](../ship-pr/SKILL.md)) merges every lens's
findings into one fix plan.

**Read-only.** Do not edit files, commit, push, or comment on GitHub. You may
read, search, run `git`, and run targeted Go tests. Do not run `xcodebuild`;
the gate already ran it and parallel builds collide.

## Lenses

| Lens | ID | Rubric | Asks |
| --- | --- | --- | --- |
| Conformance and regression | `CON` | [`lenses/conformance.md`](lenses/conformance.md) | Every AC met, nothing out of scope, nothing that worked is broken, tests prove it |
| Maintainability | `MNT` | [`lenses/maintainability.md`](lenses/maintainability.md) | Is there a shape that fits the codebase better and costs the next change less? |
| User experience | `UX` | [`lenses/ux.md`](lenses/ux.md) | Copy, accessibility, workflow, and error handling |
| Security | `SEC` | [`lenses/security.md`](lenses/security.md) | Can a file, project, or value the app did not create make it misbehave? |
| Performance | `PRF` | [`lenses/performance.md`](lenses/performance.md) | Does it stay fast on a real project and off the main thread? |

## Inputs (from the orchestrator)

- **Lens**: one of the five, or **Security and performance** together. In
  that case one reviewer reads both rubrics, works each one's **How to work**,
  and numbers findings `SEC-` or `PRF-` by the rubric they come from. A finding
  that is both (an unbounded decode, an unbounded list loaded from the catalog)
  is written once, under the lens with the higher severity, and names the other.
- **Plan**: `docs/deployment-plan/<spike>/prs/S#-NN-<slug>.md`.
- **Base** and **head**: commits. The change is `git diff <base>...<head>`.
- **Round**: 1, or 2 with the **ship log** (`S#-NN-<slug>.ship.md`) holding the
  round-1 findings and how each was answered.

## How to work

1. Read the rubric for your lens, top to bottom. Its **Read first** list is
   your context; read it.
2. Read the plan: User stories, Scope, Decisions, Codebase fit, Design.
3. Read the diff, then the code around it. Every lens needs the neighbors, not
   only the changed lines.
4. Work the rubric's **How to work** and **Look for**. Search before every claim.
5. Write findings in the format below, most severe first. Then a verdict.

**Round 2.** Your scope is the fix commits (`git diff <round-1 head>...<head>`)
and the round-1 findings for your lens. For each round-1 blocker or major
marked Fix now, say **Resolved** or **Not resolved** with the evidence (the
orchestrator checks the minors). Raise a new
finding only if it is a **blocker**.

**Stay in your lens.** If you notice something serious in another lens, add it
at the end under **Outside my lens**, one line each. The orchestrator routes it.

## Finding format

```markdown
### UX-3 · major · One-line title
**Where:** `path/to/file.swift:42` (and others)
**Plan:** AC-1.2 / D-3 / Out line / — (what in the plan it touches)
**Problem:** what goes wrong, for whom, when.
**Evidence:** the code, the trace from input to failure, the command and its output, or the failing test.
**Fix:** what to change. Name the existing code to reuse when there is one.
**Fit:** this PR | follow-up (preparatory PR, idea file)
**Confidence:** high | medium | low
```

**Nits do not use this format.** List them last, under **Nits**, one line
each: `MNT-5 · path/to/file.go:42 · what to change`. Do not write Evidence,
Fix, or Fit for a nit. A nit that needs evidence to be believed is a minor.

| Severity | Means | Loop effect |
| --- | --- | --- |
| **blocker** | Must not merge: an AC not met, a regression, data loss, a security hole, a hang, debt expensive to undo, a defect in code this PR adds that a later PR in the map will reach | Fixed before the PR is done |
| **major** | Should be fixed in this PR: a real problem on a real path, cheap enough here | Fixed this PR, or explicitly deferred by the orchestrator |
| **minor** | Worth doing; not worth a round | Fixed only if the fix plan already touches the file; else follow-up |
| **nit** | Taste | One line under **Nits**; logged, never fixed in the loop |

Each rubric has its own severity table; use it.

### Rate the code as merged

Severity measures the defect in the code once it is on `main`, not how soon
something calls it. Three framings never lower a severity:

- **"Nothing calls it yet."** A later PR in the map that will call this code
  (an FFI field, a renderer, a handler) makes the defect reachable. When the
  defect is in code this PR adds or changes, it is fixed **in this PR**, at
  the severity it will have once that caller lands. Do not write
  `Fit: follow-up (S#-NN)` for it. Naming the PR that exposes a hole is not a
  reason to ship the hole.
- **"The plan said so."** A Decision that spells out the weaker shape does not
  make the debt smaller. Rate it by the debt, and say which Decision it
  contradicts so the orchestrator can take it to the user.
- **"It is small."** A ten-line private copy of an existing mechanism is still
  a second mechanism.

`Fit: follow-up` is for problems in code this PR does not add, change, or
make worse.

**Evidence rules.** A finding with no evidence is not a finding. Conformance and
regression findings carry the failing test (see the conformance rubric). A
low-confidence finding says what would confirm it.

**Verdict:** `Clean` (no blocker or major) or `Changes needed`, and one line on
what matters most.

## Area audit (no PR)

The same lenses audit an area of the app on demand — what the old app-health
review did. Use it periodically, after a run of AI-assisted PRs, or when an
area feels heavy.

1. **Scope.** A path set (`macos/App/Features/Sources`, `core/search`) or the
   whole app. Default lenses: all five; the user can pick.
2. **Per lens**, sample instead of reading everything: the high-churn files
   (`git log --since=… --name-only`), the area's entry points, and its tests.
   Conformance in an audit checks the area against its domain doc and its
   skills instead of a plan. Maintainability Q10 checks that every skill and
   rule touching the area still matches the code.
3. **Report in chat** (a file only if the user asks), grouped by lens, then one
   **action list**: each item the smallest meaningful PR — one concern, one
   done state, reviewable without a design debate. Title it like a PR subject.
   Mark each item High / Medium / Low: High for security, data loss, a broken
   primary flow, or guidance that trains agents wrong; Low for polish.
4. Do not fix anything. The user picks items; they go to the map with
   [`plan-spike`](../plan-spike/SKILL.md) or to `docs/ideas/`.
