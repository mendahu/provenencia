# Spike 12 — Source structure

## Status

**Open.** Slices 1–4 have landed (S12-01 – S12-19). Slices 5–7, the fixes from the [app-health audit](app-health-audit.md) of 2026-10-27, are scoped, and slices 5b and 6b, the fixes from the [second audit](app-health-audit-2.md) of 2026-10-28, follow slices 5 and 6; the spike closes when they land. PR map and working rules: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md). Spike 11 is still open alongside it; the PRs that touch the same Promote and graph-loading code wait on its PRs, named in the map.

Every Source has had exactly one flat Evidence graph. This spike lets a big Source (a census, a book, a directory) hold several isolated graphs, each with its own Artifacts and promoted on its own, and adds Copy Source so an item inside a publication (an obituary) starts from a sibling's metadata.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | Goal, working rules, PR map, order |
| [**Source structure**](source-structure.md) | Design for Copy Source and multiple Evidence graphs: the two tools, data model, rules, where graphs show, migration |
| [**App-health audit**](app-health-audit.md) | The 2026-10-27 audit findings behind slices 5–7: where, evidence and fix for each, and the Low items not scheduled |
| [**App-health audit 2**](app-health-audit-2.md) | The 2026-10-28 audit findings behind slices 5b and 6b, the decisions taken on them, the Low items not scheduled, and the review gates each class slipped past |
| [**Completed**](completed.md) | Finished steps |
| [Design briefs](design/) | Claude Design, written just before the PR they gate |
| [PR plans](prs/) | One per PR, written just before the PR |

## Not this spike

- **Nested Sources, derivation links, named templates, per-graph credibility, layered citation output:** [`ideas/source-structure-later.md`](../../ideas/source-structure-later.md).
- **Focus views inside one large graph:** [`ideas/archive/graph-focus-views.md`](../../ideas/archive/graph-focus-views.md) (archived 2026-10-24).
- **`source` Subjects, `mentions` and `remark`:** [`ideas/source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md). Graph isolation changes its assumptions; re-think it when it is pulled.
- **Moving a Subject between graphs, or splitting a graph:** refused; graphs are fully isolated.
- **Matching changes:** Spike 11.
