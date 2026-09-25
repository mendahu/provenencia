# Artifact thumbnails from PDF

**Status:** idea only — not roadmapped. Descoped from Spike 8 (**S8-02** retired; do not reuse the id).

PDF Artifacts keep the **file-type glyph**. Image Files still get a raster derivative. A Source cover stays a user-pinned raster (or the type icon) — so a PDF still cannot be cover until an engine renderer exists.

## Why it left Spike 8

Telling two PDFs apart on Artifact rows is real, but not worth a client-only path. macOS **PDFKit** → `file_derivatives` would leave Windows on the glyph and put generation in the frontend. Thumbnails are catalog derivatives; they belong in **Go** if they come back, same as JPEG/PNG today.

## If revived

- Page 1 (media box) only unless a later story picks a cover page.
- Same `ThumbnailSpec()` as images: JPEG, longest edge ≤ 256. Persist as unaudited `file_derivatives.thumbnail`. No schema / `cover_mode` change — a raster unlocks the existing pin.
- Encrypted, empty, corrupt, or over the source-byte budget → skip, keep the glyph. No partial page.
- **Generate in `core/derivatives`**, not Swift. No `PutFileThumbnail` client bytes path.
- Out: Quick Look, video / Office posters, user-chosen thumb page, PDF OCR.

Shipped glyph rules: [`archive/file-type-fallback-thumbnails.md`](archive/file-type-fallback-thumbnails.md). Image pipeline: [`core/derivatives`](../../core/derivatives/).
