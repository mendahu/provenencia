# Artifact thumbnails from PDFKit

**Status:** scoped as Spike 8 **S8-02**. Promoted from `docs/ideas/`.

**Implements as:** PR **S8-02** (no design brief — existing `PVThumbnail` raster vs glyph).

## Problem

`derivatives.EnsureThumbnail` only rasterizes common **image** MIME types. PDF Artifacts skip and fall back to a **file-type glyph** on Artifact rows ([archived fallback](../../ideas/archive/file-type-fallback-thumbnails.md)). Researchers cannot tell two PDFs apart. Source cover pins already require a **raster**, so PDF Artifacts cannot become Source cover.

## Locked for S8-02

| Decision | Choice |
| --- | --- |
| Which page | **Page 1** (media box) only. User-picked cover page is later. |
| Size / format | Same as `ThumbnailSpec()`: JPEG, longest edge ≤ **256**. See [`image-optimization.md`](../../ideas/archive/image-optimization.md). |
| Where it runs | **macOS PDFKit** renders; writes the usual thumbnail **File + `file_derivatives`** row the catalog already expects. Go `EnsureThumbnail` **does not** grow a PDF decoder (Windows / engine stay skip → glyph until a later renderer). |
| Bad PDFs | Encrypted, empty, corrupt, or over existing source-byte budget → **skip**, keep glyph. No partial page. |
| Cover pin | A successful PDF raster **unlocks** existing “use as Source cover” without schema / `cover_mode` changes. |
| UI | No new chrome. Artifact rows and Source cover already paint `thumbnailRelPath` or the glyph. |

## Shape

1. Resolve the File under the project (same path rules as ingest / viewers).
2. Open with **PDFKit** (`PDFDocument` / first page) — same stack as S7-06 `ArtifactViewerModel` / `PDFPage.thumbnail(of:for:)`.
3. Render page 1 to JPEG at thumbnail size **in memory**.
4. Persist via a small FFI (extend `EnsureFileThumbnail` or a `PutFileThumbnail` that accepts JPEG bytes) so list / pin plumbing is unchanged.

Generation stays unaudited (same as today’s image thumbs). No Quick Look. No Office / video posters.

## Out

- Replacing MIME glyphs for types we still cannot raster
- Video poster frames / Office preview
- Changing `cover_mode` / pin schema
- OCR of PDF pages (S8-01 is image-only)
- User-chosen thumbnail page

## Related

- [`core/derivatives`](../../../core/derivatives/) (`EnsureThumbnail` skips `application/pdf` today)
- [`docs/artifact-file-storage.md`](../../artifact-file-storage.md)
- Archived: [`file-type-fallback-thumbnails.md`](../../ideas/archive/file-type-fallback-thumbnails.md)
- Spike 7 PDF paint: S7-06 `ArtifactViewerModel` / PDFKit
- Plan step: [`deployment-plan.md`](deployment-plan.md) **S8-02**
