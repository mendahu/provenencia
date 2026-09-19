# Spike 7 — Citations, Observations, and the citation composer

## Status

**Open.** Checklist and PR sequence: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md) (none yet).

Stand up the Interpretation **Citation → Observation** pipeline on the Evidence graph: **Subject fields** editor, NameValue, image/PDF citation composer as a **navigable place**, cited property rows on cards, and durable connect macros. Authoritative design: [`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §4–§6 / slices 3–7. Canvas inherits from [Spike 6](../archive/spike-6/) (Go).

> **UI dogfood is inverted:** Add property + card chrome first, then a **thin** composer (text cite → card grows), then viewers / locators / NameValue. See [`deployment-plan.md`](deployment-plan.md) § Incremental UI dogfood.

> **Composer is Option B:** navigate away from the graph to a first-class workspace place (viewer \| form). Not an in-window modal over the canvas, and not a companion `NSWindow`.

> **Subject types are not user-editable.** The seven seeded kinds (person / event / place / bridges / source) stay product-seeded with first-class graph plumbing. No Subject types CatalogVocabulary UI in this spike (S7-D1 / S7-04 descoped). Behavior (capabilities, locked bindings, connect matrix, **Property term** sets) is declared in one **Interpretation subject registry** in S7-01 / **S7-01b** — the plugin extension point later.

> **Property terms (S7-01b):** kind/edge Properties (`event_type`, `role`, `relationship_type`) use `value_type = term` + `property_terms` — not free text. No Event types / Roles admin destinations; composer picker ± Add custom for user terms.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | PR sequence, design gates, dogfood bar, scope |
| [**Completed**](completed.md) | Finished steps (empty until landings) |
| [Design briefs](design/) | Claude Design — open S7-D2…S7-D5 (S7-D1 descoped) |

## Relationship to Spike 5 / 6 / later

Spike 5 shipped subjects, positions, Subject types seed, and Sources-family stubs. Spike 6 shipped the canvas prototype (provisional connect). Spike 7 is slices **3–7** of the design note collapsed into one spike: Subject **fields** editor (not types), **Property terms**, artifact viewer + Citations, Observations (six value types including NameValue as its **own design stream**), and durable connect.

**Later (not this spike):** Source-page `mentions` / `remark`, audio/video composers, Citation pinning across graph edits, conflicted/negated visual language, unplaced tray / minimap, removing the Subject types sidebar stub if desired.

## Out of scope (for this spike)

- **Subject types** CatalogVocabulary editor / user-defined Subject types (product-seeded + first-class only)
- Event types / Roles / Relationship-types CatalogVocabulary destinations (Property terms use composer picker ± Add custom)
- In-window composer modal over the graph; companion `NSWindow`
- `real` / `boolean` Property value types
- Audio / video / QuickLook-as-composer
- Conclusion `name_format` profiles
- Citation pinning across successive graph edits
- Auto-layout, unplaced tray, minimap
- Product SemVer bump for docs-only planning (bump only if cutting a release)
