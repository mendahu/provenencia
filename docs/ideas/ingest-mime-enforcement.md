# Ingest MIME / file-type enforcement

**Status:** shipped — policy in [`core/ingest/mediatypes`](../../core/ingest/mediatypes); macOS fail-fast + modal drag-drop on Add Artifact / Add File. Related UI: [`file-type-fallback-thumbnails.md`](file-type-fallback-thumbnails.md). Domain: [`source-layer-data-model.md`](../source-layer-data-model.md) §6; ingest: `core/ingest`.

## What shipped

1. **Go allowlist** (`mediatypes.Resolve`) after `http.DetectContentType`. Extension fallback when sniff is `application/octet-stream`; when sniff is `text/plain`, extension may upgrade to `text/csv` / `text/markdown`.
2. **Allowed text / docs:** `text/plain`, `text/csv`, `text/markdown`, Word (`application/msword`, OOXML `.docx`), plus images, PDF, common audio/video. Excel/PowerPoint remain rejected.
3. **Family-specific codes** (`ingest.unsupported_office`, `…_archive`, `…_executable`, `ingest.unidentified`, `ingest.too_large`, `ingest.empty`, path errors, etc.).
4. **macOS fail-fast** — open-panel UTTypes + `IngestMediaPolicy.validate` (no FFI) before Create/Attach; same taxonomy Callouts.
5. **Drag-and-drop** on both ingest modals (`IngestFileDropRow`) sharing `applyPickedFile`.

Already-ingested files remain readable. Raster thumb allowlist stays separate and narrower. Office formats (docx/xlsx/…) remain out of scope.
