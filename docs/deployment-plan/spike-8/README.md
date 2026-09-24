# Spike 8 — Pause and refine (data entry)

## Status

**Open.** Checklist and PR sequence: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md).

A pause after Spike 7: the evidence path works, but **entering Sources and filling an Evidence graph is tedious**. This spike amalgamates small, mostly unrelated improvements to that flow — not a new layer. Graph chrome in **S8-D3** / **S8-06**: conflict + negated badges, jump to the Source page, richer bridge sentences, **Add property** on bridges. Incomplete-bridge chrome and collapse/expand are **descoped**. Tray, pinning, and Source-page → graph stay out unless added.

Stories land one at a time. **Auto Transcribe** (Vision on **images** only). **PDF Artifact thumbnails** (first-page raster so rows and Source covers can tell PDFs apart). **PDF Find + select + paste transcription** (live PDFKit page; I-beam default; no Vision on PDF). **Graph visual enhancements** (badges, Source jump, bridge copy + Add property).

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

**Later (not this spike unless a story is added):** Citation pinning across subjects, graph+composer rethink, Foundation Models draft cards, PDF OCR / Vision on page rasters, `text_quote` locators, Source-page Find, Source-page → graph, user-picked PDF cover page, audio/video, leftover honesty (denied-line drawing, filters, undo), tray. Incomplete-bridge chrome and collapse/expand are descoped (not later stories).

## Out of scope (for this spike)

- New catalog tables / migrations for OCR or thumbs (reuse `file_derivatives`)
- Auto-filling Observations, subjects, or connect macros from OCR
- Apple Intelligence / PCC
- Product SemVer bump for docs-only planning (bump only if cutting a release)
