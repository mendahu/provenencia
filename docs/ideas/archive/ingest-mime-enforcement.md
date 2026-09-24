# Ingest MIME / file-type enforcement (archived)

**Done.** Policy: `core/ingest/mediatypes`. Domain: [`source-layer-data-model.md`](../../source-layer-data-model.md) §6.

## Decisions

- Go allowlist after sniff (`http.DetectContentType`), with a narrow extension fallback.
- Allowed: common images, PDF, text/csv/markdown, Word, common audio/video. Excel / PowerPoint / archives / executables stay rejected.
- Fail-fast on macOS (UTTypes + `IngestMediaPolicy`) before FFI. Already-ingested files remain readable.
- Raster thumb allowlist is separate and narrower than ingest.
