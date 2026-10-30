# Spike 14 — Odds and ends

## Status

**Open.** Batch 1 (Event and Place pictures) cut 2026-10-27; S14-01 and S14-08 have landed. Batch 3 came from Spike 11, and Batches 4 and 5 from Spike 12, when those spikes closed (2026-10-29). PR map: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md).

The first odds-and-ends spike (working rule "Even spikes are odds and ends"). It takes small features, UX refinements and fixes in batches, each a slice with its own Check. Batch 1 gives Events and Places a Picture Property, a gallery and a chosen thumbnail (parity with a Person's portrait), and makes depiction drawing rectangle-only. Batch 2 raises `segment.OtherSupport` so a thin name plus kinship clears Weak on Promote (S14-08). Batch 6 adds the Wide width class (1600pt pane) and a Page columns kit component, designed in [S14-D1](design/archive/S14-D1-wide-layout.md).

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | Goal, working rules, PR map, order |
| [**Event and Place pictures**](event-place-pictures.md) | Design for batch 1. PRs are in the map |
| [**Oversized JPEG thumbnails**](oversized-jpeg-thumbnails.md) | Design for S14-23 (batch 1, from Spike 12) |
| [**App-health audit**](app-health-audit.md) | The 2026-10-27 findings behind batch 5's `A-n` items |
| [**App-health audit 2**](app-health-audit-2.md) | The 2026-10-28 findings behind batch 5's `B-n` items |
| [**Completed**](completed.md) | Finished steps |
| [**Design briefs**](design/README.md) | Claude Design briefs, written just before the PR they gate |
| PR plans (`prs/`) | One per PR, written just before the PR |

## Not this spike

Large new features: those go to the next odd-numbered spike. For batch 1: signature and the Spike 10 audit leftovers not taken into S14-01 and S14-03 stay in [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md). The event picture in the Person Timeline is Spike 13's (R-26).
