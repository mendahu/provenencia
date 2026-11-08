# Spike 16 — Odds and ends

## Status

**Open.** Batch 1 (the Person page audit fixes) is cut; no PR has landed. PR map: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md).

The second odds-and-ends spike ([odd and even spikes](../README.md#odd-and-even-spikes)). It takes small fixes, UX refinements and follow-ups in batches, each a slice with its own Check. Batch 1 fixes what the 2026-11-08 audit of Spike 13's Person page found: a Timeline that puts a missing death after a burial, Person reads a crafted project can stall, a Timeline that builds every event, stale or silently failed data, a record links diagram that hides an uncounted read, and section switches, removals and folds that VoiceOver and the keyboard lose.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | Goal, working rules, PR map, order |
| [**Person page audit**](person-page-audit.md) | The 2026-11-08 findings behind Batch 1's `P-n` items, with evidence and failing tests |
| [**Completed**](completed.md) | Finished steps |
| [**Design briefs**](design/README.md) | Claude Design briefs, written just before the PR they gate |
| PR plans (`prs/`) | One per PR, written just before the PR |

## Not this spike

Large new features: those go to the next odd-numbered spike. For Batch 1: speed work on the Person page that real data does not yet need (participants sent once per read, a record links memo, memoised section content) is in [`ideas/person-page-performance.md`](../../ideas/person-page-performance.md); the audit's low-priority Person page items are in [`ideas/person-page-followups.md`](../../ideas/person-page-followups.md), [`ideas/structural-debt.md`](../../ideas/structural-debt.md) and [`ideas/share-packages.md`](../../ideas/share-packages.md). Event and Place pages adopting sections stay with their redesign.
