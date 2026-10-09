# Spike 10 — Claude Design briefs

**UI in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**. Workflow: [`add-design-brief`](../../../../.cursor/skills/add-design-brief/SKILL.md).

Design track and gating: [`../deployment-plan.md`](../deployment-plan.md#design-track). **One brief per view.** A PR that changes two views waits on two briefs. A later PR on a view that already has a brief does not get another one. Spike overview: [`../README.md`](../README.md).

Brief files are written **just before** the PR they gate, not up front. This index is the schedule.

## Scheduled

| Step | View | Gates | Notes |
| --- | --- | --- | --- |
| S10-D1 | Adaptive choice field (kit) | **S10-14** | Component board, not a workspace place. Composer adoption is S10-D3 / S10-16 |
| S10-D2 | Create Source dialog | **S10-15** | Field order only |
| S10-D3 | Citation composer | **S10-16** | Observation rows, add-at-top, focus, Auto Transcribe. Also draws the depiction row **S10-34** fills |
| S10-D4 | Name parts modal | **S10-17** | Split, fixed-width type control. Design after **S10-29** so `patronymic` is in the list |
| S10-D5 | Source page — Artifacts | **S10-18** | Drop target and open-on-collapsed-row |
| S10-D6 | Evidence graph | **S10-18** (header open) | Also specifies box-select, which ships in **S10-20**. Hit boxes, zoom, canvas growth, and camera restore are not this board |
| S10-D7 | Sidebar | **S10-19** | Subject marks for People, Events, Places, including the new Event mark. The toolbar section icon is the same slot and ships in S10-19 |
| S10-D8 | Sources list | **S10-22** | Last updated on the row |
| S10-D9 | Persons list | **S10-22** | Last updated, same cell as Events and Places. The existing thumb becomes a likeness crop in **S10-35**, then the preferred portrait in **S10-39** |
| S10-D10 | Events list | **S10-22** | |
| S10-D11 | Places list | **S10-22** | |
| S10-D12 | Person detail | **S10-23** | Sex / gender (`man`, `woman`, `non_binary`, extensible), the likeness gallery **S10-35** fills, and the preferred-portrait control **S10-39** fills. Design all of that before S10-23 |
| S10-D13 | Promote page | **S10-36** | Evidence sheet shows two crops |

## Done

None yet.
