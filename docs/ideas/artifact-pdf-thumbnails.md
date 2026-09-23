# Artifact thumbnails from PDFKit (and richer file rasters)

**Status:** idea only — not roadmapped.

## Problem

`derivatives.EnsureThumbnail` only rasterizes common **image** MIME types. PDF Artifacts skip cleanly and fall back to a **file-type glyph** on Artifact rows ([archived fallback idea](archive/file-type-fallback-thumbnails.md)). That is honest but weak for scanned certificates and multi-page registers — researchers cannot tell two PDFs apart at a glance.

Source cover pins already require a **raster** thumbnail, so PDF Artifacts cannot become Source cover today.

## Idea

Extend thumbnail generation so **PDF** (and maybe later other document types) produce a real raster derivative:

1. Resolve the File under the project via the same path rules as ingest / viewers.
2. Open with **PDFKit** (`PDFDocument` / first page).
3. Render page 1 (media box) to a bitmap at thumbnail size — same idea as today’s composer page raster / `PDFPage.thumbnail(of:for:)`.
4. Write the usual thumbnail derivative and reuse existing list / pin plumbing.

No Quick Look dependency required for the PDF path if PDFKit is enough; keep skipping exotic Office/video until there is a clear renderer.

## Why it fits

Artifact rows and optional Source covers become visually specific. Composer and Source Artifacts already open PDFs via PDFKit in **S7-06** — thumbnail generation is the same stack, run at ingest / ensure time instead of UI paint time.

## Open questions

- First page only vs user-picked cover page later?
- Max pixel size / DPI vs existing image thumbnail policy ([`image-optimization.md`](archive/image-optimization.md) if still relevant).
- Where to run: Go `derivatives` (needs a PDF renderer on the engine side) vs macOS-only ensure that writes the derivative File the catalog already expects?
- Encrypted / empty / huge PDFs — skip vs partial?
- Does a PDF raster unlock “Use as thumbnail” for Source cover without new schema?

## Explicitly out of this note

- Replacing MIME glyphs for non-renderable types
- Video poster frames / Office preview
- Changing cover_mode / pin schema

## Related

- [`core/derivatives`](../../core/derivatives/) (`EnsureThumbnail` skips non-images today)
- [`artifact-file-storage.md`](../artifact-file-storage.md)
- Archived: [`archive/file-type-fallback-thumbnails.md`](archive/file-type-fallback-thumbnails.md)
- Spike 7 PDF paint: **S7-06** `ArtifactViewerModel` / PDFKit
