# Spike 8 — Pause and refine (data entry)

## Status

**Open.** Checklist and PR sequence: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md).

A pause after Spike 7: the evidence path works, but **entering Sources and filling an Evidence graph is tedious**. This spike amalgamates small, mostly unrelated improvements to that flow — not a new layer. A thin **conflict badge** on repeated Properties is pulled in as graph chrome (**S8-D3** / **S8-06**); the rest of the honesty/polish slice (negated lines, tray, pinning) stays out unless added.

Stories land one at a time. **Auto Transcribe** (Vision on **images** only). **PDF Artifact thumbnails** (first-page raster so rows and Source covers can tell PDFs apart). **PDF Find + select + paste transcription** (live PDFKit page; I-beam default; no Vision on PDF). **Graph visual enhancements** (conflict badge first; more items join that PR).

> **Do not invent Observations.** Transcription is the reading. The researcher still edits and Save still writes the Citation.

> **Stay on macOS 14.** On-device Vision only. Foundation Models / Private Cloud Compute / raising the deployment target are out unless a later story explicitly takes them.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | PR sequence, design gates, dogfood bar, scope |
| [**Completed**](completed.md) | Finished steps |
| [Design briefs](design/) | Claude Design — **S8-D1**, **S8-D2**, **S8-D3** open (**S8-02** has no board) |
| [PDF thumbnails](artifact-pdf-thumbnails.md) | Scoped note for **S8-02** (promoted from ideas) |
| [PDF Find](pdf-text-find.md) | Scoped note for **S8-D2** / **S8-03…S8-05** (promoted from ideas) |

## Relationship to Spike 7 / dogfood

Spike 7 shipped the composer place, locators, and durable connect. Dogfood then showed that a single newspaper notice still means a lot of typing. Background (OCR APIs, page-size honesty, later LLM extract): [`docs/dogfood/ux.md`](../../dogfood/ux.md).

**Later (not this spike unless a story is added):** Citation pinning across subjects, graph+composer rethink, Foundation Models draft cards, PDF OCR / Vision on page rasters, `text_quote` locators, Source-page Find, user-picked PDF cover page, audio/video, full honesty language (negated lines, incomplete bridges, filters, undo), tray.

## Out of scope (for this spike)

- New catalog tables / migrations for OCR or thumbs (reuse `file_derivatives`)
- Auto-filling Observations, subjects, or connect macros from OCR
- Apple Intelligence / PCC
- Product SemVer bump for docs-only planning (bump only if cutting a release)
