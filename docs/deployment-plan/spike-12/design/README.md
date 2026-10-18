# Spike 12 — Claude Design briefs

**UI in this spike is designed in Claude Design before it is implemented.** Hand open briefs over **one at a time**. Workflow: [`add-design-brief`](../../../../.agents/skills/add-design-brief/SKILL.md).

Design track and gating: [`../deployment-plan.md`](../deployment-plan.md#design-track). **A brief covers one view. A view may have more than one brief.** Brief files are written **just before** the PR they gate, not up front. This index is the schedule.

## Scheduled

| Step | View | Gates | Notes |
| --- | --- | --- | --- |
| S12-D1 | Copy Source form | **S12-02** | Title, type dropdown, part checkboxes (one per Artifact); entry points on the Source page and Sources list context menu ([design](../source-structure.md#copy-source)) |
| S12-D2 | Source page — Graphs section | **S12-11** | One row per graph with Open and delete; replaces the header jump |
| S12-D3 | Graph form | **S12-12** | Create and edit: name, description, Artifact checkboxes |
| S12-D4 | Source page — Artifacts section graph links | **S12-13** | Checkboxes in the add dialog and card edit |
| S12-D5 | Sources list — graph accordion | **S12-14** | Expander per row, per-graph progress and Open |
