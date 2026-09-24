# PDF Find + text selection (composer)

**Status:** scoped as Spike 8 **S8-D2** / **S8-03…S8-05**. Promoted from `docs/ideas/`.

## Problem

S7-06 paints PDF pages as a **raster** in `ArtifactMediaViewport` and **click-drags to pan**. There is no text layer in the viewport, so researchers cannot Find a surname in a long register or select a line to transcribe. Image Auto Transcribe (**S8-01**) is Vision on photos; born-digital PDFs already have text — OCR is the wrong tool.

## Product (this spike)

1. **Find** — field on the PDF tool strip (`ArtifactViewerToolChrome`). Type a keyword, Find (and next/previous). Jump the page chrome to the hit and **highlight** it. Navigation only — not a locator type.
2. **Select text** — default pointer on PDF is an **I-beam**. Drag selects `PDFSelection`. Pan is no longer the unnamed default (board picks explicit hand tool and/or modifier-drag).
3. **Paste transcription from selection** — control above / on the transcription field copies the current PDF selection into `transcription` (replace-confirm if non-empty, same honesty as Auto Transcribe).

Composer-only. Image Artifacts unchanged (still raster + region + **S8-01**). Image-only PDFs (no text layer): Find and select **honest empty** — do **not** OCR in this slice.

## Why the PRs are ordered this way

Find and paste both need a **live PDFKit page** (`PDFView` / equivalent), not `PDFPage` flattened to `NSImage`. Region overlay (S7-07) today maps onto the raster document. **S8-03** remounts PDF paint onto PDFKit and remaps pan / select / region coordinates. **S8-04** (Find) and **S8-05** (paste) cannot ship on the raster viewer.

**S8-02** (page-1 thumbnail) also uses PDFKit but writes a derivative at ingest — it does **not** unblock the composer viewport. **S8-01** is the image twin of S8-05; prefer S8-05 after S8-01 so one transcription action row.

## Locked

| Decision | Choice |
| --- | --- |
| Search API | `PDFDocument.findString` / `beginFindString` → `PDFSelection` |
| Default gesture (PDF) | Text selection. Pan is explicit. |
| Region tools | Stay; exclusive with select (armed region tool wins). Find stays available. |
| No text layer | Disable Find / paste; short reason. No Vision fallback. |
| Source-page viewer | Out. Composer only. |
| `text_quote` locator | Out. Hits do not write locator JSON. |

## Out

- PDF OCR / Vision on page rasters
- Full in-document Find UI on Source page
- Seeding `text_quote` from hits
- Changing locator schema

## Related

- Plan: [`deployment-plan.md`](deployment-plan.md) **S8-D2**, **S8-03…S8-05**
- Brief: [`design/S8-D2-pdf-text-find.md`](design/S8-D2-pdf-text-find.md)
- Spike 7 raster viewer: archive **S7-06** / **S7-07**
- Image OCR twin: **S8-D1** / **S8-01**
- PDFKit: `PDFDocument.findString(_:withOptions:)`
