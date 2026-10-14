# Spike 10 — Claude Design briefs

**UI in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**. Workflow: [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md).

Design track and gating: [`../deployment-plan.md`](../deployment-plan.md#design-track). **A brief covers one view. A view may have more than one brief.** A PR that changes two views waits on a brief for each. Spike overview: [`../README.md`](../README.md).

Brief files are written **just before** the PR they gate, not up front. This index is the schedule.

## Scheduled

| Step | View | Gates | Notes |
| --- | --- | --- | --- |
| [S10-D6](S10-D6-evidence-graph.md) | Evidence graph | **S10-20** | File select and Open (system app), plus box-select. Also specifies Unpromote on the promoted card, which ships in **S10-44**. Hit boxes, the tiled grid, canvas growth, and camera restore are not this board |
| S10-D8 | Sources list | **S10-41** | The page: last updated, evidence-status filter, text search, persisted sort and filters. Session scroll ships in the same PR and has no chrome of its own. Persons, Events, and Places do not get those controls |
| S10-D9 | Persons list | **S10-22** | Last updated, same cell as Events and Places. The existing thumb becomes a likeness crop in **S10-35**, then the preferred portrait in **S10-39** |
| S10-D10 | Events list | **S10-22** | |
| S10-D11 | Places list | **S10-22** | |
| S10-D12 | Person detail | **S10-23** | Sex / gender (`man`, `woman`, `non_binary`, extensible), the likeness gallery **S10-35** fills, and the preferred-portrait control **S10-39** fills. Design all of that before S10-23 |
| S10-D13 | Promote page | **S10-36** | Evidence sheet shows two crops |
| S10-D15 | Citation composer — depiction row | **S10-34** | Own board, written just before S10-34. After the observation-list rethink |

## Descoped

S10-D1, the adaptive choice field, was descoped 2026-10-12 with S10-14. Term fields keep the control they have. The brief is not kept.

S10-D2, the Create Source dialog, is not drawn. Name, then type, is a field swap, so S10-15 stands on its own.

S10-D14, the zoomed-out card, was descoped 2026-10-11 with S10-46. The full card stays at every zoom. The brief is not kept.

## Done

| Step | View | Gates | Notes |
| --- | --- | --- | --- |
| S10-D3 | Citation composer | **S10-16** | Done 2026-10-12. [S10-D3-citation-composer.md](archive/S10-D3-citation-composer.md) |
| S10-D4 | Name parts modal | **S10-17** | Done 2026-10-13. [S10-D4-name-parts-modal.md](archive/S10-D4-name-parts-modal.md) |
| S10-D5 | Source page — Artifacts | **S10-18** | Done 2026-10-14. [S10-D5-source-artifacts.md](archive/S10-D5-source-artifacts.md) |
| S10-D16 | Design system — Event mark | **S10-19** | Done 2026-10-14. [S10-D16-event-mark.md](archive/S10-D16-event-mark.md) |
| S10-D7 | Workspace shell | **S10-19** | Done 2026-10-14. [S10-D7-sidebar-marks.md](archive/S10-D7-sidebar-marks.md) |
