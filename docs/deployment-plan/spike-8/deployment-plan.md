# Deployment Plan — Spike 8

Pause-and-refine: make **Source → Evidence graph** data entry cheaper. Stories are independent improvements to that flow, not one schema epic. Authoritative composer/locator: [Spike 7 archive](../archive/spike-7/). Dogfood seed: [`docs/dogfood/ux.md`](../../dogfood/ux.md).

## Status

**Open.** Landings go in [`completed.md`](completed.md). More stories will be added under the same spike.

> **Goal of this spike:** cut the tedium of entering real research without reopening the Interpretation model. Slices land independently. Later stories join this plan as they are scoped.

## Goal (dogfood bar)

Grow this list as stories land. **By spike close**, every checked story below must be true in the app.

1. **Auto Transcribe** — in the citation composer, on an **image** Artifact, a control fills **transcription** from Vision OCR of the image (or the region polygon when one is set). The researcher can edit and Save as today. Full-image / oversized jobs warn and can still proceed. No Observation writes. **PDF:** this button stays disabled (text-layer path is bar items 3–4).
2. **PDF Artifact thumbnails** — PDF Artifacts show a **first-page raster** on Artifact rows (glyph only if render skips). A PDF with a raster can be pinned as Source cover. Notes: [`artifact-pdf-thumbnails.md`](artifact-pdf-thumbnails.md).
3. **PDF Find** — on a PDF in the composer, a Find field on the viewer tool strip jumps to a keyword hit with a highlight. Image-only PDFs (no text layer) fail honestly. Notes: [`pdf-text-find.md`](pdf-text-find.md).
4. **PDF select + paste transcription** — default PDF pointer is text select (pan is explicit). **Paste transcription from selection** fills the transcription field. No Vision on PDF pages.
5. **Graph visual enhancements** — conflict + negated row badges; always-on **jump to the Source page**; cited **bridge sentences** prefer endpoint `name` / `event_type` / `toponym`, then working label; **Add property** on bridge cards (extra non-edge rows). More items may join **S8-D3** / **S8-06**.
6. **Source page enhancements** — the Source detail page has an **Open Evidence graph** control for the same Source (disabled with no Artifact). More items may join **S8-D4** / **S8-07**.

Further bar items: TBD (additional data-entry stories).

## Design track

**Composer, graph, and Source-page chrome is designed in Claude Design before the matching UI PRs.** PDF thumbs reuse shipped `PVThumbnail` — no board. Briefs: [`design/`](design/).

| Step | Brief | Covers | Gates |
| --- | --- | --- | --- |
| **S8-D1** | Auto Transcribe in the composer | Button, progress, replace confirm, large-page warning + proceed, failure copy | **S8-01** |
| **S8-D2** | PDF Find + select + paste | Tool-strip Find; I-beam vs pan; paste-from-selection vs Auto Transcribe row | **S8-03**, **S8-04**, **S8-05** |
| **S8-D3** | Evidence graph visual enhancements | Conflict + negated; Source-page jump; richer bridge sentences; Add property on bridges | **S8-06** |
| **S8-D4** | Source page enhancements | Jump to Evidence graph; more page items join this brief | **S8-07** |

## PR sequence

```text
   design                         build
─────────────               ──────────────────────────────────────────

S8-D1  Auto Transcribe UI
  │
  └────── gates ──────────▶ S8-01  Vision + crop + fill transcription
                              │     (images only; parallel with S8-02)

S8-02  PDF first-page thumbs ──     (no design gate; does not unblock viewer)

S8-D2  PDF Find / select / paste
  │
  └────── gates ──────────▶ S8-03  PDFKit live page + I-beam default + pan mode
                              │     (MUST precede Find and paste)
                              ├──────▶ S8-04  Find field + highlight + page jump
                              └──────▶ S8-05  Paste transcription from selection
                                              (after S8-03; prefer after S8-01
                                               so one transcription action row)

S8-D3  Graph card visuals
  │
  └────── gates ──────────▶ S8-06  Graph chrome (badges, Source jump, bridge copy + Add property)
                              │     (parallel; no composer / PDF dependency)

S8-D4  Source page enhancements
  │
  └────── gates ──────────▶ S8-07  Open Evidence graph from the Source page
                              │     (parallel; pair with S8-06 graph → page)
                              │
                            S8-99  Dogfood close / docs
```

**Order notes**

- **S8-03** remounts PDF from raster → PDFKit. Find and paste cannot ship on `ArtifactMediaViewport` bitmaps.
- **S8-04** and **S8-05** are parallel after **S8-03**.
- **S8-01** / **S8-02** do not block **S8-D2**. **S8-05** should follow **S8-01** when both touch the transcription `PVField`.
- **S8-02** uses PDFKit only to write a thumbnail derivative — not the composer viewport.
- **S8-D3** / **S8-06** are independent of OCR and PDF remount. Freeze the **S8-06** bundle on the brief before that PR starts.
- **S8-D4** / **S8-07** are independent of OCR, PDF remount, and **S8-06**. Freeze the **S8-07** bundle on the brief before that PR starts. The two jumps (graph ⇄ page) should use the same location helpers and product name.

---

## Checklist

- [ ] S8-D1 — Design: Auto Transcribe in the citation composer → [`completed.md`](completed.md)
- [ ] S8-01 — Vision OCR + Auto Transcribe button → [`completed.md`](completed.md)
- [ ] S8-02 — PDF first-page Artifact thumbnails → [`completed.md`](completed.md)
- [ ] S8-D2 — Design: PDF Find, text selection, paste transcription → [`completed.md`](completed.md)
- [ ] S8-03 — PDFKit live viewer + I-beam default → [`completed.md`](completed.md)
- [ ] S8-04 — PDF Find in the tool strip → [`completed.md`](completed.md)
- [ ] S8-05 — Paste transcription from PDF selection → [`completed.md`](completed.md)
- [ ] S8-D3 — Design: Evidence graph visual enhancements → [`completed.md`](completed.md)
- [ ] S8-06 — Graph visual enhancements (badges, Source jump, bridge copy + Add property) → [`completed.md`](completed.md)
- [ ] S8-D4 — Design: Source page enhancements → [`completed.md`](completed.md)
- [ ] S8-07 — Source page enhancements (Evidence graph jump + brief bundle) → [`completed.md`](completed.md)
- [ ] S8-99 — Dogfood close / docs (after later stories, or when we choose to close)

---

## S8-D1 — Design: Auto Transcribe in the citation composer

Claude Design board for the **transcription** field: Auto Transcribe control, in-progress state, replace confirm, large-page / slow-job warning that can still proceed, and failure/empty states. Brief: [`design/S8-D1-auto-transcribe.md`](design/S8-D1-auto-transcribe.md). Gates **S8-01**.

Does **not** design Observation auto-fill, LLM extract, PDF OCR, or PDF Find (**S8-D2**).

---

## S8-01 — PR: Vision OCR + Auto Transcribe

On-device Vision (`VNRecognizeTextRequest`) fills the composer **transcription** textarea for **image** Artifacts only. Crop in memory from the loaded `NSImage` and the region locator when present. No temp file, no catalog write until the researcher Saves. PDF / audio / video: do not run Vision.

| | |
| --- | --- |
| **In** | Protocol-shaped OCR seam (tests do not call Vision); `CGImage` from the already-loaded **image**; bounding-box crop (optional mask later); Auto Transcribe control per **S8-D1**; replace confirm if transcription is non-empty; preflight warn on artifact-only / huge bitmap / dense-image heuristics, with proceed; L10n; skip PDF / audio / video / no image. |
| **Out** | PDF OCR; PDF page raster → Vision; PDF Find / select / paste (**S8-03…S8-05**); Foundation Models; writing Observations; persisted crop objects; `RecognizeDocumentsRequest` (macOS 26); raising the deployment target. |
| **Testable** | Fake recognizer fills / fails / empty; crop uses region vs full page; preflight flags large page; replace does not overwrite without confirm; button disabled while running. |
| **Depends on** | **S8-D1**. Shipped composer + locators (S7-08 / S7-06 / S7-07). |

---

## S8-02 — PR: PDF first-page thumbnails

macOS **PDFKit** renders page 1 of a PDF Artifact File to the same JPEG thumbnail spec as images (longest edge ≤ 256). Persist as the existing unaudited `file_derivatives` thumbnail so Artifact rows and Source cover pins just work. Go `EnsureThumbnail` still **skips** PDF decode — no engine PDF renderer in this PR.

| | |
| --- | --- |
| **In** | Page 1 only; JPEG / 256-edge; PDFKit (same as S7-06 viewer raster); FFI to store bytes as the thumbnail File + link; skip encrypted / empty / corrupt / over-budget → keep file-type glyph; Source cover pin works once a raster exists (no schema change); existing `EnsureFileThumbnail` / list lookup. |
| **Out** | User-picked cover page; video / Office posters; Quick Look; Go PDF library; `cover_mode` / pin schema; OCR; new `PVThumbnail` chrome. |
| **Testable** | Image thumbs unchanged; PDF with a first page gets a `thumbnailRelPath`; bad PDF stays skipped + glyph; pin-as-cover accepts a PDF that has a raster; idempotent ensure. |
| **Depends on** | Shipped `derivatives.EnsureThumbnail` + `PVThumbnail` glyph fallback. **Not** S8-D1 / S8-01. |

Notes: [`artifact-pdf-thumbnails.md`](artifact-pdf-thumbnails.md).

---

## S8-D2 — Design: PDF Find, text selection, paste transcription

Claude Design board for the PDF **tool strip** (Find), **cursors** (I-beam default vs pan), and transcription **Paste from selection**. Brief: [`design/S8-D2-pdf-text-find.md`](design/S8-D2-pdf-text-find.md). Gates **S8-03**, **S8-04**, **S8-05**.

Does **not** design image OCR, PDF thumbnails, or `text_quote` locators.

---

## S8-03 — PR: PDFKit live page + I-beam default

Replace the composer PDF **raster** (`displayImage` / `ArtifactMediaViewport`) with a **PDFKit-backed** page so `PDFSelection` exists. Default drag **selects text**. Pan is an explicit hand tool and/or modifier (per **S8-D2**). Region overlay (S7-07) still draws in page space. Image viewer unchanged.

| | |
| --- | --- |
| **In** | Live `PDFDocument` / `PDFView` (or equivalent) for PDF Artifacts; I-beam default; pan mode; cursors; region tools exclusive with select; zoom/page chrome still work; no-text-layer is paintable (select does nothing useful). |
| **Out** | Find UI (**S8-04**); paste button (**S8-05**); Vision; changing image pan; Source-page viewer. |
| **Testable** | PDF path no longer depends on `displayImage` for hit-testing text; image path unchanged; region draft still normalizes; pan mode still moves the page; selecting text does not pan. |
| **Depends on** | **S8-D2**. Shipped S7-06 / S7-07. **Not** S8-01 / S8-02. |

This is the load-bearing remount. Do not start S8-04 / S8-05 until it lands.

---

## S8-04 — PR: PDF Find

Find field on `ArtifactViewerToolChrome` (PDF only). `PDFDocument.findString` → highlight + page jump + next/previous.

| | |
| --- | --- |
| **In** | Keyword field; Find / next / prev; current-hit highlight; sync `model.page`; no-match and no-text-layer copy; L10n; disable while a region tool is drawing if the board says so (Find itself may stay). |
| **Out** | Source-page Find; `text_quote` locators; OCR fallback; image Find. |
| **Testable** | Fake/document fixture: hit changes page; wrap/stop per board; empty query / no hits; hidden on image Artifacts. |
| **Depends on** | **S8-D2**, **S8-03**. |

---

## S8-05 — PR: Paste transcription from PDF selection

Transcription-row control for PDF: copy current `PDFSelection` string into `transcription`. Replace confirm if the field is non-empty (same as **S8-01**).

| | |
| --- | --- |
| **In** | Button/label per **S8-D2**; disabled with no selection / no text layer / `inert`; replace confirm; does not write Observations; image row stays Auto Transcribe. |
| **Out** | Vision on PDF; auto-running paste; locator writes. |
| **Testable** | Selection string fills the field; empty selection disabled; replace does not overwrite without confirm; image Artifact does not show an enabled Paste. |
| **Depends on** | **S8-D2**, **S8-03**. **Prefer after S8-01** so the transcription `PVField` action row is designed once. |

---

## S8-D3 — Design: Evidence graph visual enhancements

Claude Design board for a **bundled** graph-chrome pass. Items: **conflict** + **negated** row badges; always-on **Source-page jump**; **bridge sentences** that prefer endpoint identity Properties; **Add property** on bridge cards. Brief: [`design/S8-D3-graph-visuals.md`](design/S8-D3-graph-visuals.md). Gates **S8-06**.

Does **not** design a composer rethink (pinning / empty Citation — dogfood) or denied-lines. Source-page → graph is **S8-D4**. **Descoped:** incomplete-bridge chrome, collapse/expand, density filters, undo, unplaced tray, minimap.

---

## S8-06 — PR: Graph visual enhancements

One Evidence graph chrome pass against **S8-D3**. Competing Observations stay as separate rows. Duplicate `propertyKey` → conflict badge on each row; `polarity = negative` → negated mark (may stack). Header (or equivalent) **opens the same Source’s detail page**. Cited bridge sentences prefer each endpoint’s identity Observation (`name` / `event_type` / `toponym`), then working label. Bridge cards get **Add property** (reuse `composerLocation(for:)`) and show extra **non-edge** Observation rows. Further items listed on the brief at PR start ship here.

| | |
| --- | --- |
| **In** | Conflict + negated per **S8-D3**; Source jump via existing `sourcePageLocation` + `go(to:)`; `EvidenceBridgeEdgeSummary` reads endpoint Observations from the snapshot; Add property + extra rows on bridges; L10n + VoiceOver; card height / hit tests. Prefer `PVBadge`. |
| **Out** | Merge / resolve; schema or FFI; denied-line drawing; Source-page → graph (**S8-07**); composer rethink (dogfood). Incomplete-bridge chrome, collapse/expand, filters, undo, tray, minimap are **descoped**. |
| **Testable** | Two `name`s → both conflict; negative singleton → negated only; jump location is `.page` for the same `sourceId` and Back returns to `.graph`; relationship sentence uses NameValue form when present and label when not; participation uses `event_type`; location uses `toponym`; bridge Add property opens composer for that bridge (not the connect Citation); extra non-edge row visible + editable; edge keys not duplicated as rows; height/a11y follow the new sentence. |
| **Depends on** | **S8-D3**. Shipped cards + snapshot + `sourceSurface`. **Not** S8-01…S8-05 / **S8-07**. |

---

## S8-D4 — Design: Source page enhancements

Claude Design board for a **bundled** Source-page chrome pass. First item: **Open Evidence graph** for this Source (disabled with no Artifact). More page items join this brief (and **S8-07**) as they are scoped. Brief: [`design/S8-D4-source-page.md`](design/S8-D4-source-page.md). Gates **S8-07**.

Does **not** design graph chrome, source-to-source commentary ([`source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md)), composer rethink, or Sources-list counts.

---

## S8-07 — PR: Source page enhancements

One Source-page chrome pass against **S8-D4**. Add a control that opens this Source’s Evidence graph (`sourceSurface: .graph`). Same `hasArtifact` rule as the list split-row. Further items listed on the brief at PR start ship here.

| | |
| --- | --- |
| **In** | Jump control per **S8-D4**; reuse `SourcesListNavigation.graphLocation` (or equivalent); disabled + reason when no Artifact; `go(to:)`; L10n + VoiceOver; any other SP items frozen on the brief. Prefer `PVButton`. |
| **Out** | Graph chrome (**S8-06**); opening the graph with zero Artifacts; source-to-source commentary ([`source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md)); composer rethink; list redesign. |
| **Testable** | Source with an Artifact → location is `.graph` for the same `sourceId`; Back returns to `.page`; no Artifact → control disabled and does not navigate. |
| **Depends on** | **S8-D4**. Shipped Source page + `sourceSurface`. **Not** S8-01…S8-06. |

---

## S8-99 — Dogfood close / docs

Honesty pass against the [goal bar](#goal-dogfood-bar) once the cluster is enough (or we stop adding stories). Record in [`completed.md`](completed.md); archive the spike. SemVer only if cutting a product release.

---

## Scope boundary

| In | Out |
| --- | --- |
| Composer transcription assist | Auto Observations / subjects / connect |
| Vision on **image** Artifacts | PDF OCR; audio / video OCR |
| In-memory crop from locator | Object-store crop files |
| Warn + proceed on large images | Hard reject / Apple “too many words” (does not exist) |
| PDF **page-1 thumbnail** via PDFKit | Go PDF decoder; user-picked thumb page |
| PDF **Find** + **select** + **paste transcription** | PDF Vision / OCR; `text_quote` locators; Source-page Find |
| Graph **conflict** + **negated** badges; Source-page jump; richer bridge sentences; **Add property** on bridges | Denied-line drawing; composer rethink / pinning (dogfood); merge/resolve; **descoped** leftovers (incomplete bridges, collapse/expand, filters, undo, tray, minimap) |
| Source page **Open Evidence graph** (more page items via **S8-D4**) | Source-to-source commentary (`mentions` / `remark`, placeholder + merge — [`source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md)); Sources-list counts |
| More data-entry stories as added | Remaining Spike 7 leftovers unless pulled in |

---

## Gotchas

1. **Transcription ≠ Observation** — OCR dumps into the citation reading only ([`interpretation-graph-ui.md`](../../ideas/interpretation-graph-ui.md) §4).
2. **Vision does not refuse a newspaper page** — it usually succeeds slowly or with junk. Large-page honesty is **our** preflight (pixels / no region / post-pass observation density), not a `VNError`.
3. **Locator y-down vs Vision ROI y-up** — crop in image pixels from [`ArtifactRegionGeometry`](../../../macos/App/Features/ArtifactViewer/ArtifactRegionGeometry.swift); do not pass a polygon into `regionOfInterest` (rect only).
4. **Crop the source raster**, not the zoomed viewport bitmap.
5. **No file middleman** — `ProjectFiles.objectURL` → image `NSImage` → `CGImage` → crop → `VNImageRequestHandler`.
6. **PDF is not an OCR input** — S8-01 disables Auto Transcribe. S8-05 pastes a PDFKit selection. Image-only PDFs get neither Vision nor fake text.
7. **Hide Vision behind a protocol** — `FakeStore` / unit tests inject a recognizer.
8. **More stories do not wait on S8-01** unless they share the composer transcription chrome.
9. **S8-02 is macOS PDFKit → catalog derivative**, not `core/derivatives` learning to parse PDF. Windows keeps the glyph until a later engine renderer.
10. **S8-03 before Find/paste** — raster `displayImage` has no `PDFSelection`. Region overlay must remount with the live page or locators break.
11. **Pan vs select** — today’s unnamed click-drag pan will fight I-beam. S8-D2 must name the pan escape (hand and/or modifier) before S8-03.
12. **Conflict is a count, not a verdict** — badge when `propertyKey` appears ≥ 2 times on that card. Values may match. Do not write a schema flag or a resolve action.
13. **Negated is polarity, not a missing line** — `polarity = negative` on the Observation. Italic-danger today is not enough; S8-D3 designs an explicit mark. Do not draw a ghost connect edge.
14. **Incomplete bridges are descoped** — Connect is atomic and the UI cannot write person-without-event. Do not add half-line chrome.
15. **Bridge nouns come from the endpoint card** — prefer that subject’s `name` / `event_type` / `toponym`, then `subjects.label`. Do not keep using only the edge row’s working-label display once a name exists.
16. **Graph → page is S8-06; page → graph is S8-07** — reuse `sourcePageLocation` and `SourcesListNavigation.graphLocation`. Same `hasArtifact` gate as the list.
17. **Collapse/expand is descoped** — do not hide the bridge sentence behind a disclosure.
18. **Bridge Add property is a new Citation** — `composerLocation(for: bridgeID)`, not `composerLocationForBridgeCitation` (that edits the connect Citation). Extra rows omit edge keys already in the sentence.
19. **Unplaced tray is descoped** — canvas create and Connect always write a position. The snapshot **omits** subjects with no row. No UI path produces a tray candidate until imports exist.
20. **Filters / undo / minimap are descoped** — density stays a dogfood note; undo can return if ⌘Z becomes a real pain; minimap was scope-creep.

---

## Definition of done

- [ ] Checklist stories complete (or explicitly descoped)
- [ ] Dogfood bar items for landed stories met
- [ ] Design briefs archived under `design/archive/`
- [ ] [`docs/deployment-plan/README.md`](../README.md) points at archive when closed
