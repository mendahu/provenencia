# Spike 12 — Source structure

## Status

**Open.** Scoped 2026-10-18; no PR has landed yet. PR map and working rules: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md). Spike 11 is still open alongside it; the PRs that touch the same Promote and graph-loading code wait on its PRs, named in the map.

Every Source has had exactly one flat Evidence graph. This spike lets a big Source (a census, a book, a directory) hold several isolated graphs, each with its own Artifacts and promoted on its own, and adds Copy Source so an item inside a publication (an obituary) starts from a sibling's metadata.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | Goal, working rules, PR map, order |
| [**Source structure**](source-structure.md) | Design for Copy Source and multiple Evidence graphs: the two tools, data model, rules, where graphs show, migration |
| [**Completed**](completed.md) | Finished steps |
| [Design briefs](design/) | Claude Design, written just before the PR they gate |
| [PR plans](prs/) | One per PR, written just before the PR |

## Not this spike

- **Nested Sources, derivation links, named templates, per-graph credibility, layered citation output:** [`ideas/source-structure-later.md`](../../ideas/source-structure-later.md).
- **Focus views inside one large graph:** [`ideas/graph-focus-views.md`](../../ideas/graph-focus-views.md).
- **`source` Subjects, `mentions` and `remark`:** [`ideas/source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md). Graph isolation changes its assumptions; re-think it when it is pulled.
- **Moving a Subject between graphs, or splitting a graph:** refused; graphs are fully isolated.
- **Matching changes:** Spike 11.
