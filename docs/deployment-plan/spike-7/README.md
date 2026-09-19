# Spike 7 — Citations, Observations, and the citation composer

## Status

**Open.** Checklist and PR sequence: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md) (none yet).

Stand up the Interpretation **Citation → Observation** pipeline on the Evidence graph: Subject vocabulary editors, NameValue, image/PDF citation composer as a **navigable place**, cited property rows on cards, and durable connect macros. Authoritative design: [`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §4–§6 / slices 3–7. Canvas inherits from [Spike 6](../archive/spike-6/) (Go).

> **Composer is Option B:** navigate away from the graph to a first-class workspace place (viewer \| form). Not an in-window modal over the canvas, and not a companion `NSWindow`.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | PR sequence, design gates, dogfood bar, scope |
| [**Completed**](completed.md) | Finished steps (empty until landings) |
| [Design briefs](design/) | Claude Design — open S7-D1…S7-D5 |

## Relationship to Spike 5 / 6 / later

Spike 5 shipped subjects, positions, Subject types seed, and Sources-family stubs. Spike 6 shipped the canvas prototype (provisional connect). Spike 7 is slices **3–7** of the design note collapsed into one spike: vocabulary editors, artifact viewer + Citations, Observations (five value types including NameValue as its **own design stream**), and durable connect.

**Later (not this spike):** Source-page `mentions` / `remark`, audio/video composers, Citation pinning across graph edits, conflicted/negated visual language, unplaced tray / minimap.

## Out of scope (for this spike)

- In-window composer modal over the graph; companion `NSWindow`
- `real` / `boolean` Property value types
- Audio / video / QuickLook-as-composer
- Conclusion `name_format` profiles
- Citation pinning across successive graph edits
- Auto-layout, unplaced tray, minimap
- Product SemVer bump for docs-only planning (bump only if cutting a release)
