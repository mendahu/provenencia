# Spike 8 — Pause and refine (data entry)

## Status

**Open.** Checklist and PR sequence: [`deployment-plan.md`](deployment-plan.md). Finished steps: [`completed.md`](completed.md).

A pause after Spike 7: the evidence path works, but **entering Sources and filling an Evidence graph is tedious**. This spike amalgamates improvements to that flow — not a new layer. **Composer rethink first** (**S8-D7** / **S8-10**): Citation as the document, multi-subject observations, reuse, empty Save. **Then composer and Connect simplification** (**S8-D8** / **S8-11**): row-level Observation commits, an unsaved-work guard, Connect finished in the composer, and lossless audit for Interpretation writes. Then Auto Transcribe and PDF Find/paste land on that form. Graph chrome in **S8-D3** / **S8-06**; Source-page chrome in **S8-D4** / **S8-07**; Sources-list chrome in **S8-D5** / **S8-08**; delete paths in **S8-D6** / **S8-09**.

Stories land one at a time. **Composer rethink** (flexible Citation document). **Auto Transcribe** (Vision on **images** only). **PDF Find + select + paste transcription** (live PDFKit page; I-beam default; no Vision on PDF). **Graph visual enhancements** (badges, Source jump, bridge copy + Add property). **Source page enhancements** (jump to Evidence graph; metadata dates as text). **Sources list refresh** (subject + observation counts on their own cache keys). **Delete paths** (counted cascade or honest refuse; no Change type).

> **Do not invent Observations.** Transcription is the reading. The researcher still edits and Save still writes the Citation.

> **Stay on macOS 14.** On-device Vision only. Foundation Models / Private Cloud Compute / raising the deployment target are out unless a later story explicitly takes them.

## Documents

| Doc | Role |
| --- | --- |
| [**Deployment plan**](deployment-plan.md) | PR sequence, design gates, dogfood bar, scope |
| [**Completed**](completed.md) | Finished steps |
| [Design briefs](design/) | Claude Design — **S8-D7** first, then **S8-D8**, then **S8-D1**…**S8-D6**. Working rules: [`design/claude-design-working-rules.md`](design/claude-design-working-rules.md) |
| [PDF Find](pdf-text-find.md) | Scoped note for **S8-D2** / **S8-03…S8-05** (promoted from ideas) |

## Relationship to Spike 7 / dogfood

Spike 7 shipped the composer place, locators, and durable connect. Dogfood then showed that a single newspaper notice still means a lot of typing. Background (OCR APIs, page-size honesty, later LLM extract): [`docs/dogfood/ux.md`](../../dogfood/ux.md).

**Later (not this spike unless a story is added):** Foundation Models draft cards, PDF OCR / Vision on page rasters, Source-page Find, user-picked PDF cover page. Parked ideas (not this spike): **source-to-source** commentary ([`source-to-source-relationships.md`](../../ideas/source-to-source-relationships.md)); **first-class Source provenance date** ([`source-provenance-date.md`](../../ideas/source-provenance-date.md)); **`text_quote` locators** ([`text-quote-locators.md`](../../ideas/text-quote-locators.md)); **audio / video Sources** ([`audio-video-sources.md`](../../ideas/audio-video-sources.md)); **PDF Artifact thumbnails** ([`artifact-pdf-thumbnails.md`](../../ideas/artifact-pdf-thumbnails.md) — engine renderer if revived, not Swift/PDFKit). QuickLook preview is **descoped**.

**Pulled from dogfood:** composer rethink — one reading × many subjects, Citation reuse / pinning, empty-Citation Save (**S8-D7** / **S8-10**).

**Descoped (not later stories):** incomplete-bridge chrome, collapse/expand, density filters (also in dogfood), undo, unplaced tray, minimap, QuickLook, Change type UI (delete + place), adopt/import subjects, removing the Subject types sidebar stub, user-minted NameValue part types, raising macOS 14 for the canvas, rich-text notes. The UI cannot write incomplete bridges or unplaced subjects; tray/minimap/adopt were import/scope-creep insurance. Conclusion / Narrative leftovers (case view, Conclusion chrome, Sameness workflow, family tree) are those layers’ docs — not this spike.

## Out of scope (for this spike)

- New catalog tables / migrations for OCR (reuse `file_derivatives`)
- PDF Artifact thumbnails (glyph stays; parked in [`artifact-pdf-thumbnails.md`](../../ideas/artifact-pdf-thumbnails.md))
- Auto-filling Observations, subjects, or connect macros from OCR
- Apple Intelligence / PCC
- Product SemVer bump for docs-only planning (bump only if cutting a release)
