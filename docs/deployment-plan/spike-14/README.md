# Spike 14 — Odds and ends

## Status

**Open.** Batches 1–6 largely landed; remaining Scoped work includes S14-20, S14-26 and **Batch 7** (dogfood UX, cut 2026-11-04) and **Batch 8** (a recorded read index for derived rows, cut 2026-11-07). PR map: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md).

The first odds-and-ends spike (working rule "Even spikes are odds and ends"). It takes small features, UX refinements and fixes in batches, each a slice with its own Check. Batch 1 gives Events and Places a Picture Property, a gallery and a chosen thumbnail (parity with a Person's portrait), and makes depiction drawing rectangle-only. Batch 2 raises `segment.OtherSupport` so a thin name plus kinship clears Weak on Promote (S14-08). Batch 6 adds the Wide width class (1600pt pane) and a Page columns kit component, designed in [S14-D1](design/archive/S14-D1-wide-layout.md). Batch 7 takes PDF page sync, polygon edit, Evidence graph polish, list/nav polish and a boolean Property value type. Batch 8 makes summary rows and search documents record what they read, so a write finds what it moved by lookup instead of a hand-written inverse.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | Goal, working rules, PR map, order |
| [**Event and Place pictures**](event-place-pictures.md) | Design for batch 1. PRs are in the map |
| [**Oversized JPEG thumbnails**](oversized-jpeg-thumbnails.md) | Design for S14-23 (batch 1, from Spike 12) |
| [**Derived rows know what they read**](derived-read-index.md) | Design for batch 8 (S14-40, S14-42) |
| [**App-health audit**](app-health-audit.md) | The 2026-10-27 findings behind batch 5's `A-n` items |
| [**App-health audit 2**](app-health-audit-2.md) | The 2026-10-28 findings behind batch 5's `B-n` items |
| [**Batch 7 dogfood notes**](ux/) | UX notes pulled into S14-28 – S14-37 |
| [**Completed**](completed.md) | Finished steps |
| [**Design briefs**](design/README.md) | Claude Design briefs, written just before the PR they gate |
| PR plans (`prs/`) | One per PR, written just before the PR |

## Not this spike

Large new features: those go to the next odd-numbered spike. For batch 1: signature and the Spike 10 audit leftovers not taken into S14-01 and S14-03 stay in [`ideas/depictions-and-likenesses.md`](../../ideas/depictions-and-likenesses.md). The event picture in the Person Timeline is Spike 13's (R-26). Evidence graph structure shortcuts (wedding / paste cluster) stay in [`ideas/evidence-graph-structure-shortcuts.md`](../../ideas/evidence-graph-structure-shortcuts.md).
