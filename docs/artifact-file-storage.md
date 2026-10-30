# Provenencia Genealogy — Artifact and File Storage

## Status

The Artifact/File storage design has been consolidated into [`source-layer-data-model.md`](source-layer-data-model.md).

That document is the authoritative reference for:

- `artifacts`
- `files`
- `file_derivatives`
- offline-first file ingestion
- content-addressed storage (a stored type off the ingest allowlist has no object path on read)
- immutable primary files
- first-attach only (no primary-File pointer-swap; better scan = new Artifact)
- generated thumbnails and other disposable derivatives

This file is retained only as a pointer so existing references do not lead to a stale duplicate schema.
