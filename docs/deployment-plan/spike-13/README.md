# Spike 13 — Person detail page

## Status

**Open.** Requirements agreed and PR map cut 2026-10-26; no PR has landed. PR map: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md).

The Person page is a stub. This spike turns it into a fixed header over five sections (Vitals, Family, Timeline, Gallery, Sources) with quick switching between them, adds a reusable provenance sheet beside every value in Vitals, and adds controls to remove or review a person's identity claims. Each view is designed in Claude Design and then implemented, one brief at a time.

**Slice 0 runs first** (added 2026-10-29): it removes Identity Claim Observation pins, Promote's evidence sheet and the engine behind them, the precursor the claims sections build on ([`conclusion-layer-data-model.md`](../../conclusion-layer-data-model.md) §5). Claim review (S13-18) waits on open question 9.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | Goal, working rules, PR map, order, open questions |
| [**Person detail**](person-detail.md) | Requirements, spike-wide rules, cross-spike dependencies, not this spike |
| [**Source provenance date**](source-provenance-date.md) | The first-class Source date the Sources section shows (moved from ideas) |
| [**Completed**](completed.md) | Finished steps |
| [Design briefs](design/) | Claude Design, written just before the PR they gate |
| PR plans (`prs/`) | One per PR, written just before the PR |

## Not this spike

Event and Place pages in the new layout, Reconciliation Claims, the comparison of a member against its handle that replaces pins ([conclusion §5.1](../../conclusion-layer-data-model.md#51-two-lenses)), editing identity claims, weak-claim alerts, event depictions, kinship beyond the nuclear family, and sibling evidence in matching. Homes: [person-detail.md § Not this spike](person-detail.md#not-this-spike).
