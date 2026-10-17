# Spike 11 — Promote Matching Refinement

## Status

**Open.** Scoped 2026-10-16 – 17; no PR has landed yet. PR map and working rules: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md). Spike 10 is still open alongside it.

Spike 9 shipped Promote graph alignment, and Spike 10 tuned its comparers. This spike rebuilds matching as one layered, registry-driven module the rest of the app borrows from, and deletes the old matching code and its duplicates as each layer lands. We are in early development, so churn is welcome: no compatibility layers, ruthless migrations.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | Goal, working rules, PR map, order |
| [**Comparison layers**](comparison-layers.md) | Design for layers 1–6 (value, property, Subject, entity and segment comparison, graph alignment), bands, lifespan, the match registry, and the reuse and retirement inventory |
| [**Claim evidence**](claim-evidence.md) | Design for the Promote sheet's pins: every member's evidence, grouped by property, lazily queried, in plain language |
| [**Connection subrows**](promote-connection-subrows.md) | Design for showing connections under the Promote rows they join, with no switch-off |
| [**Completed**](completed.md) | Finished steps |
| [Design briefs](design/) | Claude Design, written just before the PR they gate |
| [PR plans](prs/) | One per PR, written just before the PR |

## Not this spike

- **The rest of the matching test suite** (scenario goldens runner, random-graph invariants, simulated world, scoreboard, labeler, fitting, dogfood export) stays in [`ideas/matching-test-suite.md`](../../ideas/matching-test-suite.md). Whether it joins this spike is decided after the re-implementation.
- **Merge-hint UI and the weak-claim review UI.** The layers make both cheap (subject compare on two profiles; `subjectprofile.Build` with `Excluding`), but the views stay in [`ideas/identity-claim-review.md`](../../ideas/identity-claim-review.md).
- **Term affinity** (birth / baptism, son / child), **Daitch–Mokotoff** ([`ideas/daitch-mokotoff.md`](../../ideas/daitch-mokotoff.md)), the **place gazetteer** ([`ideas/place-gazetteer-service.md`](../../ideas/place-gazetteer-service.md)), and **maiden names in the conclusion**.
- **Cardinality beyond birth and death**, two-hop lifespan rules, damped multi-hop credit, and a coarse misspelling shortlist key: start simple, revisit with evidence.
- **Per-project registry overrides** and researcher-facing tuning.
