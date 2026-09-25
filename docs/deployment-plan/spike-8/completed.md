# Spike 8 — Completed steps

Finished Spike 8 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S8-NN`, `S8-DN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| S8-D7 | Design | Citation composer rethink — Layout A + 1500pt two-column; C/D rejected |
| S8-10 | PR | Flexible composer: Citation is the document; identity, reuse, empty Save |

## Steps

### S8-D7 — Design: Citation composer rethink

**Board pick:** Layout **A** — viewer (~60%) beside a 520pt stacked form (document line, transcription, description, Observation stack). At window ≥ **1500pt** the form widens to 760–880pt and splits **reading | observation stack**. Layout **C** (subject headings) and **D** (four columns) stay rejected.

Agreed on the board, not a second mode: 1 Artifact hides the Artifact chip; 2–3 use a ghost `PVButton` + `PVContextMenuPanel`; Citation identity is New + `CIT-…` with observation count and transcription snippet. Empty Save is quiet, not a blocker. Connect keeps two locked endpoint rows.

Brief archived: [`design/archive/S8-D7-composer-rethink.md`](design/archive/S8-D7-composer-rethink.md). Next composer chrome is **S8-D1** (Auto Transcribe).

### S8-10 — Flexible citation composer

Shipped the S8-D7 form. The composer is no longer subject-locked: every Observation on the Citation is listed, each row names a Subject, and Add property can reuse an existing reading on the Artifact.

**What shipped**

- In-form Artifact + Citation identity; Artifact pre-screen removed
- `ListByArtifact` + observation counts (`METHOD_LIST_CITATIONS_BY_ARTIFACT` = 74)
- Empty Citation create/update (schema already allowed it; engine no longer requires ≥1 Observation)
- Dirty Artifact change on a saved Citation confirms abandon
- Inline text / integer / term; name and date stay on the existing form dialog
- After Save, return to the graph with a workspace toast

**What stayed out**

- Vision / OCR, PDF remount / Find / paste (**S8-01+**)
- Minting Subjects, project-wide Citation search, `WorkspaceLocation.observationId`
- New kit primitive, four-list browser
