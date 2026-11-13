# Spike 15 — Claude Design briefs

**UI in this spike is designed in Claude Design before it is implemented.** Workflow: [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md).

Briefs are written one at a time, just before the PR they gate. Design track and gating: [`../deployment-plan.md`](../deployment-plan.md#design-track). One brief, one PR: a brief gates exactly one PR and draws only what it ships (S15-D1 and S15-D14 predate the rule and stay as written).

## Scheduled

| Brief | View | Board | Gates |
| --- | --- | --- | --- |
| [S15-D1](S15-D1-source-page.md) | Source page: the date, the Private switch, Page columns, an Artifact's metadata board, the citation preview, Move on Artifacts and graphs. The citation preview (SD-11 – SD-13) is not built (user, 2026-11-13) | [validated 2026-11-10](https://claude.ai/artifact/443NPNYmGErAmxgV1fBsFK) | S15-03, S15-09 (landed [#802](https://github.com/mendahu/provenencia/pull/802)), S15-25 |
| S15-D3 | Person page Sources section: each Source's date | — | S15-05 |
| S15-D4 | Adding an Artifact: the copy fields | Descoped 2026-11-12: folded into S15-D14 | — |
| [S15-D14](S15-D14-source-page-artifacts.md) | Source page Artifacts: the list with each Artifact's copy details read-only and a thumbnail of at most 80pt; the Artifact sheet in edit and add mode (replaces S15-D1 SD-9, SD-10 and S15-D4) | — (not drawn) | S15-34, S15-35, S15-10, S15-25 |
| S15-D5 | Evidence graph form: the item's metadata board, needed fields flagged | — | S15-11 |
| S15-D6 | Source type page: suggestions at three levels | — | S15-12 |
| S15-D7 | New Source type dialog: Copy from | — | S15-18 |
| S15-D8 | Citation composer: the printed page, unknown or supplied, flagged when needed | — | S15-13 |
| S15-D9 | Provenance sheet: formatted citations, grouping, Copy citation, the on-open loading and failed states | — | S15-21 |
| S15-D10 | Move confirm: the bundle, the target, the credibility note | — | S15-25 |
| S15-D11 | Event page: value rows open the provenance sheet; the Why accordion removed | — | S15-31 |
| S15-D12 | Place page: value rows open the provenance sheet; the Why accordion removed | — | S15-31 |
| S15-D15 | Person page Sources section: Citations grouped under one citation | — | S15-22 |
| S15-D16 | Source type page: the three formats, Insert field, live preview, warnings | — | S15-17 |
| S15-D17 | Provenance sheet: the record table's fold | — | S15-27 |
| S15-D18 | Provenance sheet: records loading on open, and their failure | — | S15-29 |
| S15-D19 | Source page: Copy source-list entry | — | S15-23 |
| S15-D20 | Source page: needed and used fields on the Source's board; the header's empty-date flag | — | S15-39 |
| S15-D21 | Artifact sheet and list: needed and used fields; the list entry's mark | — | S15-40 |
| S15-D22 | Metadata board row: Unknown and Supplied | — | S15-41 |
| S15-D23 | Source date dialog: No date and Supplied | — | S15-42 |

## Archived

| Brief | View | Board | Built by |
| --- | --- | --- | --- |
| [S15-D13](archive/S15-D13-copy-source-dialog.md) | Copy Source dialog: a Published date tick | [agreed 2026-11-10](https://claude.ai/artifact/443NPNYmGErAmxgV1fBsFK#artboard-32c863db011e) | S15-32 (#778) |
| [S15-D2](archive/S15-D2-sources-list.md) | Sources list: the date on rows and as a sort | [agreed 2026-11-10](https://claude.ai/artifact/JDm51QdorUvY79EFoci4cm) | S15-04 (#774) |
