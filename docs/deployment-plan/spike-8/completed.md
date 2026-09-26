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
| S8-02 | Descoped | PDF page-1 Artifact thumbs — not this spike; engine-only if revived |
| S8-03 | PR | PDFKit live page + I-beam; scroll pans like Preview |
| S8-04 | PR | PDF Find on the composer strip; wrap, highlight, page jump |
| S8-D2 | Design | PDF Find + I-beam select + paste transcription from selection |
| S8-05 | PR | Paste PDF I-beam selection into citation transcription |
| S8-D3 | Design | Evidence graph visuals — ochre conflict bracket + Not; Source jump; identity nouns; Add property on bridges |
| S8-06 | PR | Graph chrome: row marks, Source-page jump, richer bridge sentences, Add property on bridges |
| S8-D4 | Design | Source page — Evidence graph jump; text metadata; delete; url links |
| S8-07 | PR | Source page jump + text metadata + delete + urlshape + credibility pin |

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

### S8-02 — PDF first-page Artifact thumbnails (descoped)

Not this spike. PDF Artifacts keep the file-type glyph. A Swift/PDFKit derivative would not work on Windows and is not the client's job. If the idea returns, generate in `core/derivatives` like image thumbs. Parked: [`artifact-pdf-thumbnails.md`](../../ideas/artifact-pdf-thumbnails.md). The **S8-02** id is retired.

### S8-03 — PDFKit live viewer + I-beam default

Composer PDF Artifacts paint a live `PDFView` instead of a flattened page raster. Drag selects text (I-beam). Trackpad, mouse wheel, and scrollbars pan, matching Preview. Region locators stay unit-square y-down on the media box. Image Artifacts keep click-drag pan.

**What shipped**

- `ArtifactPDFViewport` snowflake: single-page `PDFView`, chrome zoom/page bind, native `PDFSelection`
- `ArtifactViewerModel` keeps `PDFDocument` and no longer writes `displayImage` for PDF
- Region overlay remounts on the live page (`mediaRectOverride`); armed tool still wins over select

**What stayed out**

- Find field / `findString` highlight (**S8-04**)
- Paste transcription from selection (**S8-05**)
- Vision / Live Text on PDF; changing image pan or wheel-zoom
- Source-page Find; archiving **S8-D2** (still gates 04 / 05)

### S8-04 — PDF Find in the tool strip

Composer PDF Artifacts get Find on the viewer strip: a trailing search button opens a 40pt row with keyword, match count, previous/next, and wrap. Hits highlight on the live `PDFView` and jump the page chrome. Scanned PDFs open the row with an honest no-text note and a disabled field.

**What shipped**

- `ArtifactViewerModel` Find session: `findString` (case-insensitive), wrap, page jump, no-match / no-text notes
- Composer strip: search `PVIconButton` (⌘F) + under-strip `PVInput` row; Done / Esc close
- `ArtifactPDFViewport` applies `highlightedSelections` and `go(to:)` for the active hit

**What stayed out**

- Paste transcription from selection (**S8-05**)
- Vision / Live Text on PDF
- Source-page Find; `text_quote` locators
- Archiving **S8-D2** (closed with **S8-05**)

### S8-D2 — Design: PDF Find, text selection, paste transcription

**Board pick:** Find is a trailing search `PVIconButton` that opens a 40pt under-strip row (wrap at ends). PDF default is I-beam select; pan is Preview-style scroll, not a hand tool. The transcription slot shows **one** fill action: image **Auto transcribe**, PDF **Paste transcription from selection**. Replace confirm when the field is non-empty. Scanned PDFs stay honest — no Vision.

Brief archived: [`design/archive/S8-D2-pdf-text-find.md`](design/archive/S8-D2-pdf-text-find.md). Shipped as **S8-03**, **S8-04**, **S8-05**.

### S8-05 — Paste transcription from PDF selection

Composer PDF Artifacts replace Auto transcribe with Paste. The I-beam `PDFSelection` fills the transcription field; a non-empty field asks first. The selection stays on the page so the researcher can check it. Image Artifacts keep Auto transcribe.

**What shipped**

- `ArtifactViewerModel` reports I-beam `currentSelection` (text + page); cleared on unload / new load
- Transcription row: one trailing `PVButton` — Paste on PDF, Auto transcribe on image
- Replace confirm via the existing `.pvConfirm(item:)` host; empty field pastes immediately
- Hints for no selection / selected lines / after paste / no text layer

**What stayed out**

- Vision / Live Text on PDF
- Source-page Find / paste
- `text_quote` locators
- Changing image Auto Transcribe or image pan

### S8-D3 — Design: Evidence graph visual enhancements

**Board pick:** ochre conflict **bracket** in an 11px gutter joining each competing `propertyKey` run (not `PVBadge`); **"Not"** in danger micro-caps before a denied value (italic-danger kept); ghost **Jump to Source page** with a trailing arrow after the muted Source title; bridge sentences prefer endpoint identity Observations then working label; cited bridges reuse primary rows + **Add property** for extra non-edge Observations; canvas legend floats bottom-left while any mark is present.

Agreed: badges are notices, not resolve actions; identity pick is first positive then first; Add property opens a new Citation (`composerLocation(for:)`), never the connect Citation.

Brief archived: [`design/archive/S8-D3-graph-visuals.md`](design/archive/S8-D3-graph-visuals.md). Shipped as **S8-06**.

### S8-06 — Graph visual enhancements

Shipped the S8-D3 graph-chrome pass. Competing Observations stay as separate rows. Conflict is a shared ochre bracket; negation is an explicit “Not”. The header always jumps to the same Source’s filing page. Cited bridges name endpoints from identity Properties and can take extra Observations.

**What shipped**

- Conflict bracket + “Not” prefix on cited rows (`EvidenceCitedPropertyMarks` / `EvidenceCitedPropertyRow`); canvas legend while any mark exists
- Always-on **Jump to Source page** via `sourcePageLocation` + `go(to:)`
- `EvidenceBridgeEdgeSummary` identity-Observation nouns (`name` / `event_type` / `toponym`), then working label
- Add property + extra non-edge rows on bridge cards (new Citation, not the connect Citation)

**What stayed out**

- Merge / resolve; denied-line drawing
- Incomplete-bridge chrome, collapse/expand, filters, undo, tray, minimap

### S8-D4 — Design: Source page enhancements

**Board pick:** keep shipped metadata quick-add / quick-edit; treat every field (including former date keys) as text; trash on saved rows; saved `url` values as underlined links; identity-header **Evidence graph** jump already drawn (secondary, branch glyph, `hasArtifact` gate). Credibility sits at the bottom of the left overview column; title pencil centers on the first line.

Agreed: no DateValue catalog chrome; no `PVLink`; host-shape validation is engine-only; delete is `ClearSourceMetadata`, not suggestion dismiss or vocabulary delete.

Brief archived: [`design/archive/S8-D4-source-page.md`](design/archive/S8-D4-source-page.md). Shipped as **S8-07**.

### S8-07 — Source page Evidence graph jump and text metadata

Shipped the S8-D4 Source-page pass. The identity header jumps to this Source’s Evidence graph. Catalog metadata is filing text again: dates are phrases, saved rows can be cleared, and `url` values are external links.

**What shipped**

- Trailing secondary **Evidence graph** jump (`SourcesListNavigation.graphLocation` + `hasArtifact`); disabled tooltip reuses the list’s needs-an-artifact copy; Back returns to the page
- Metadata stays shipped quick-edit; former date keys are text; trash → `ClearSourceMetadata` (suggestion returns to the dashed list; extras vanish)
- Saved `url` values are `textLink` + underline; click opens the default browser (`https://` prefix when the typed string has no scheme)
- Host-shape lives only in `core/urlshape`; `SetSourceMetadata` returns `sourcemetadata.invalid` on junk / `file:` / no host. The client puts that on the value field; I/O stays on `pageError`
- Catalog `000028`: flatten `date_value_id` into `value_text`, drop the column, `data_type` is `text` \| `url`; seeds and search projector follow
- Credibility pinned to the bottom of the left overview column; title pencil mid-first-line (`PVInlineEdit.RestingAlignment.firstLineCenter`)

**What stayed out**

- In-app browser / `PVLink`; a Swift URL parser or `ValidateURL` RPC
- DateValue on Observations / the composer (unchanged)
- First-class Source provenance date or list sort
- Redesigning the jump or the rest of the identity header
