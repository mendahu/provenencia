# Spike 8 — Completed steps

Finished Spike 8 work kept for history. Spike overview: [`README.md`](README.md). Plan and sequence: [`deployment-plan.md`](deployment-plan.md).

IDs stay stable (`S8-NN`, `S8-DN`). Do not renumber when moving steps here.

## Index

| Step | Kind | One-liner |
| --- | --- | --- |
| S8-D7 | Design | Citation composer rethink — Layout A + 1500pt two-column; C/D rejected |
| S8-10 | PR | Flexible composer: Citation is the document; identity, reuse, empty Save |
| S8-D8 | Design | Composer + Connect simplification — row-level commits, one-row connection, citation wording |
| S8-11 | PR | Row-level composer commits, Connect in the composer, lossless audit |
| S8-D1 | Design | Auto Transcribe — trailing secondary button; combined replace + whole-page confirm; image-only |
| S8-01 | PR | Vision OCR fills image-citation transcription; reusable `Features/OCR` module |

## Steps

### S8-D7 — Design: Citation composer rethink

**Board pick:** Layout **A** — viewer (~60%) beside a 520pt stacked form (document line, transcription, description, Observation stack). At window ≥ **1500pt** the form widens to 760–880pt and splits **reading | observation stack**. Layout **C** (subject headings) and **D** (four columns) stay rejected.

Agreed on the board, not a second mode: 1 Artifact hides the Artifact chip; 2–3 use a ghost `PVButton` + `PVContextMenuPanel`; Citation identity is New + `CIT-…` with observation count and transcription snippet. Empty Save is quiet, not a blocker. Connect keeps two locked endpoint rows.

Brief archived: [`design/archive/S8-D7-composer-rethink.md`](design/archive/S8-D7-composer-rethink.md). Next composer chrome was **S8-D8**.

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

### S8-D8 — Design: Composer and Connect simplification

**Board pick:** keep Layout A + the 1500pt two-column form. Replace the footer Save with per-row Save / Revert / Delete plus **Save citation**. Connect skips the graph sheet and finishes as one compact composer row (read-only endpoints + role / relationship type). Location connections have no term ("No role or type"; Save enabled immediately; never touched). Bridges are named by their computed sentence. New person / event / place from the row subject picker. Copy is citation wording, not "reading".

Agreed: a connection is one unit — wrong endpoint means discard or delete the bridge and connect again. Observation refs (`OBS-…`) show on persisted rows; saved connections show the bridge ref (`CPA-…` / `CRL-…` / `CLO-…`). Unsaved-work guard on every in-app `go(to:)`.

Brief archived: [`design/archive/S8-D8-composer-connect-simplification.md`](design/archive/S8-D8-composer-connect-simplification.md). Auto Transcribe shipped as **S8-01**. Next composer chrome is **S8-D2** (PDF Find / paste).

### S8-11 — Composer and Connect simplification + Interpretation write integrity

Shipped the S8-D8 save model and the S8-10 review leftovers that still applied: no omission-delete, a navigation leave guard, attach-to-existing-Citation Connect, full-state create/delete audit, and engine-locked edge Observations.

**What shipped**

- Observations commit one at a time (`UpdateObservation` / `DeleteObservation` / `AddObservationsToCitation`). The bundled update RPC 71 is reserved and gone.
- First commit (row Save, Save citation, or Save connection) creates the Citation from the current citation draft. After that, citation fields and rows write independently.
- Connect is one atomic `create_cited_bridge` with server midpoint and endpoint binding. A pending location connection has no term and is never touched.
- Edge Observations are immutable (`observations.edge_locked`). The composer shows one connection row, not edge rows.
- Creates record every column; `audit.Record` rejects zero-change revisions. `CreateSubject` writes optional placement in the same transaction.
- `WorkspaceLeaveGuard` holds `go(to:)` / back / forward / index when the composer has unsaved document work or a touched connection.
- Bridges are named by `EvidenceBridgeEdgeSummary.sentence(for:in:)` (working label → type + ref; unreadable edges → stored label → kind phrase · ref).

**What stayed out**

- Autosave, ⌘Z, app-quit / window-close / project-switch guards
- Subject / bridge / whole-Citation delete cascades (**S8-09**)
- Identity-Observation nouns on bridges (**S8-06**)
- Auto Transcribe / PDF Find (**S8-01+**)

### S8-D1 — Design: Auto Transcribe

**Board pick:** trailing secondary `sm` **Auto transcribe** in the transcription label row (beside Uncertain). Replace and whole-page share **one** confirm when both apply. The textarea, Save citation, identity menus, and locator tools lock while running. Image Artifacts only — never a PDF page raster, audio, video, or missing file.

Agreed: italic `PVField` hint says what will be read or why the button is disabled; fail/empty keeps prior text and shows a compact warning callout; Uncertain stays manual; no new kit primitive.

Brief archived: [`design/archive/S8-D1-auto-transcribe.md`](design/archive/S8-D1-auto-transcribe.md). Shipped as **S8-01**.

### S8-01 — Auto Transcribe for image citations

Vision OCR fills the citation transcription field on **image** Artifacts. A locator region crops in memory; no region reads the whole image (and warns when that page is oversized). The researcher edits and **Save citation** writes the Citation. Filling the field dirties the S8-11 leave guard.

**What shipped**

- Reusable `Features/OCR` module: `OCREngine` protocol, `OCRImage` crop/preflight, `VisionOCREngine` (`VNRecognizeTextRequest`)
- Composer wires image-only enablement, locator → crop, replace / whole-page confirms, running lock, fail callout
- Tests inject a fake engine; PDF / audio / missing-file paths stay disabled

**What stayed out**

- PDF OCR / PDF page rasters; PDF Find / paste (**S8-D2** / **S8-03…S8-05**)
- Observation extract; persisted crops; Live Text overlay
- `PVAutoTranscribe` kit primitive; raising macOS 14; product version bump
