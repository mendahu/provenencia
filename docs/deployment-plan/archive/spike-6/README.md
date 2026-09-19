# Spike 6 — Evidence graph canvas (UI risk spike)

## Status

**Done / Go.** Spike archived after S6-D1…S6-05. Finished steps: [`completed.md`](completed.md). Canvas prototype passed the dogfood bar; geometry reuse contract lives in [`GraphCanvas/README.md`](../../../../macos/App/Features/GraphCanvas/README.md).

Stand up a Source-scoped Evidence graph **prototype** and decide whether the spatial canvas is buildable and pleasant on this stack. Authoritative design: [`interpretation-graph-ui.md`](../../../../ideas/interpretation-graph-ui.md) §7 / §11.3 / §13. Foundation inherited from [Spike 5](../spike-5/).

> **No schema work.** Spike 5 already shipped subjects, positions, FFI, and the graph place stub. This spike replaced the stub with canvas UI.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | PR sequence, design gates, dogfood bar, go/no-go |
| [**Completed**](completed.md) | Finished steps (S6-D1…S6-05) |
| [Design briefs](design/) | Claude Design — archived briefs in [`design/archive/`](design/archive/) |

## Relationship to Spike 5 / later

Spike 5 proved the data rail and left an Evidence graph stub. Spike 6 is slice 2 of the design note (**bubbles**), plus a **connect UI prototype** so relationship lines are validated as interaction risk — not as Observation macros.

**Later (not this spike):** Citations, artifact viewer, Observations / typed values, real bridge macros + disambiguation, Subject vocabulary editors, Source-page graph button, honesty/polish states.

## Out of scope (for this spike)

- Migrations / new catalog tables
- Citation composer, artifact viewer, locators
- Observation writes and property vocabulary
- Full connect macros (disambiguation, pinned Citation, audit-heavy subgraph)
- Auto-layout, minimap, multi-select, edge routing polish
- Raising the macOS deployment target
- Product SemVer bump for a prototype spike (bump only if cutting a release)
