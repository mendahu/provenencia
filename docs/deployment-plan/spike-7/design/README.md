# Spike 7 — Claude Design briefs

**All UI in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**.

PR sequence and gating: [`../deployment-plan.md`](../deployment-plan.md). Spike overview: [`../README.md`](../README.md). Design rationale: [`interpretation-graph-ui.md`](../../../ideas/interpretation-graph-ui.md).

## Open

| Step | Brief | Feeds | Notes |
| --- | --- | --- | --- |
| S7-D2 | [`S7-D2-subject-fields.md`](S7-D2-subject-fields.md) | PR S7-05 | Subject fields (Properties + bindings to **seeded** Subject types) |
| S7-D3 | [`S7-D3-evidence-graph-updates.md`](S7-D3-evidence-graph-updates.md) | PRs S7-09, S7-10 | Graph: Add property, cited rows, connect handoff — **not** the composer |
| S7-D4 | [`S7-D4-citation-composer.md`](S7-D4-citation-composer.md) | PR S7-08 | Citation composer **place** (Option B); breadcrumbs / history — hosts NameValue, does not design it |
| S7-D5 | [`S7-D5-name-value-editor.md`](S7-D5-name-value-editor.md) | PR S7-02b | NameValue reusable modal (DateValue twin); separate stream from composer |

## Descoped

| Step | Brief | Notes |
| --- | --- | --- |
| S7-D1 | [`archive/S7-D1-subject-types.md`](archive/S7-D1-subject-types.md) | Subject types are product-seeded + first-class plumbing — **no** user CatalogVocabulary editor |

## Completed

_None yet._

## How to use

1. Open the Provenencia Claude Design project / design-system bundle (`macos/App/DesignSystem/README.md`).
2. Paste **one** open brief as the prompt for a new board or flow.
3. Keep the shipped visual language. Prefer existing `PV*` components; no new design-system primitives unless raised as a finding.
4. When the board is done, archive the brief under `archive/` and write up [`../completed.md`](../completed.md).

## Shared product facts (all briefs)

- Offline-first macOS genealogy app inside a local `*.provenencia` project.
- **Product IA:** Sources is primary work. Source types / Source fields / Subject fields are nested config under Sources. **Subject types** stay seeded (person / event / place / bridges / source) with first-class UI — not a researcher-editable vocabulary. Spike 5 may still show a Subject types stub in the rail; do not design an editor for it.
- **Evidence graph** is the product name for the Source-scoped canvas (engine: Interpretation layer — Citations / Observations / Subjects).
- **Subject fields** map to `properties` (+ bindings). Do **not** use "claim."
- Candidate refs: `CPR-…` etc. vs concluded `PER-…`.
- **Citation composer is a navigable place** (Option B): leave the graph, full-window viewer\|form, Back returns. Not a sheet over the canvas.
- Value types in this spike: **text**, **integer**, **date**, **name**, **subject** only (no `real` / `boolean`).
- **NameValue** has its own brief (**S7-D5**) — not designed inside the composer board.
- Media in composer: **image** and **PDF** only (audio/video later).
- Extend: Subject fields stub (S5-D3), Evidence graph cards (S6-D1/D2), Source types/fields CatalogVocabulary (Spike 2), DateValue editor (`Features/Dates/`).
