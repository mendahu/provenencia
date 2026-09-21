# Spike 7 — Claude Design briefs

**All UI in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**.

PR sequence and gating: [`../deployment-plan.md`](../deployment-plan.md). Spike overview: [`../README.md`](../README.md). Design rationale: [`interpretation-graph-ui.md`](../../../ideas/interpretation-graph-ui.md).

## Open

| Step | Brief | Feeds | Notes |
| --- | --- | --- | --- |
| S7-D8 | [`S7-D8-pvcallout-actions.md`](S7-D8-pvcallout-actions.md) | PR S7-14 | Kit Callout actions slot; after **S7-12**, before **S7-09** |
| S7-D3 | [`S7-D3-evidence-graph-updates.md`](S7-D3-evidence-graph-updates.md) | PRs S7-09, S7-10 | Graph first: Add property + card growth (**S7-09** before thick composer); prefer **S7-14** landed for message-center CTA |
| S7-D4 | [`S7-D4-citation-composer.md`](S7-D4-citation-composer.md) | PR S7-08 | Thin composer first (form + text); viewers/locators later |
| S7-D5 | [`S7-D5-name-value-editor.md`](S7-D5-name-value-editor.md) | PR S7-02b | NameValue modal — late fill-in after thin composer; not on path to S7-05 |
| S7-D7 | [`S7-D7-card-component.md`](S7-D7-card-component.md) | PR S7-13 | Card kit page from shipped `PVCard`; after **S7-10**, before dogfood close |

## Descoped

| Step | Brief | Notes |
| --- | --- | --- |
| S7-D1 | [`archive/S7-D1-subject-types.md`](archive/S7-D1-subject-types.md) | Subject types are product-seeded + first-class plumbing — **no** user CatalogVocabulary editor |

## Completed

| Step | Brief | Feeds | Notes |
| --- | --- | --- | --- |
| S7-D2 | [`archive/S7-D2-subject-fields.md`](archive/S7-D2-subject-fields.md) · [addendum](archive/S7-D2-subject-fields-addendum-property-terms.md) | PR S7-05 | Type strip over property table + inspector; five create value_types (**not** `term`); locked bindings as lock boxes |
| S7-D6 | [`archive/S7-D6-curated-marks.md`](archive/S7-D6-curated-marks.md) | PR S7-12 | → `Recipes/Marks/` (`file_*` / `type_*` / `subject_*` incl. source); tint API; before **S7-09** |

## How to use

1. Open the Provenencia Claude Design project / design-system bundle (`macos/App/DesignSystem/README.md`).
2. Paste **one** open brief as the prompt for a new board or flow.
3. Keep the shipped visual language. Prefer existing `PV*` components; no new design-system primitives unless raised as a finding.
4. Every open brief must include a **UI building-block inventory** (§ layered as components / recipes / snowflakes per [`docs/design-system-layers.md`](../../../design-system-layers.md)): each control lists layer, status (Ship / Extend / New / Retire), and repo home. **S7-D3** §9 is the template; **S7-D6** §7 is the design-system consolidation example.
5. When the board is done, archive the brief under `archive/` and write up [`../completed.md`](../completed.md).

## Shared product facts (all briefs)

- Offline-first macOS genealogy app inside a local `*.provenencia` project.
- **Product IA:** Sources is primary work. Source types / Source fields / Subject fields are nested config under Sources. **Subject types** stay seeded (person / event / place / bridges / source) with first-class UI — not a researcher-editable vocabulary. Spike 5 may still show a Subject types stub in the rail; do not design an editor for it.
- **Evidence graph** is the product name for the Source-scoped canvas (engine: Interpretation layer — Citations / Observations / Subjects).
- **Subject fields** map to `properties` (+ bindings). Do **not** use "claim."
- **Interpretation subject registry** (S7-01 / S7-01b) is the SoT for type capabilities, locked bindings, connect rules, and **Property term** sets — UI must not hard-code type keys or invent Event types / Roles places.
- Candidate refs: `CPR-…` etc. vs concluded `PER-…`.
- **Citation composer is a navigable place** (Option B): leave the graph, full-window viewer\|form, Back returns. Not a sheet over the canvas.
- Value types in this spike: **text**, **integer**, **date**, **name**, **subject**, **`term`** (schema). Kind/edge Properties use `term` via **registry only** — Subject fields create UI does **not** offer `term`.
- **NameValue** has its own brief (**S7-D5**) — not designed inside the composer board.
- **Term picker** (search product terms + Add custom / rename / delete user *term rows*) belongs in the composer board (**S7-D4**), not as a Subject fields CatalogVocabulary.
- Media in composer: **image** and **PDF** only (audio/video later).
- Extend: Subject fields stub (S5-D3), Evidence graph cards (S6-D1/D2), DateValue editor (`Features/Dates/`). Source types/fields are **contrast** for Subject fields — do not copy that chrome for S7-D2.
