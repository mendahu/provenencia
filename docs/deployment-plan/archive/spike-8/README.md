# Spike 8 — Pause and refine (data entry)

**Done.** Pause-and-refine made Source → Evidence graph entry cheaper: composer rethink + row commits, image OCR, PDF Find/paste, graph/page/list chrome, and named delete. Live models: [`interpretation-layer-data-model.md`](../../../interpretation-layer-data-model.md), [`catalog-deletes.md`](../../../catalog-deletes.md). Dogfood leftovers: [`docs/dogfood/ux.md`](../../../dogfood/ux.md).

## Decisions

- **Citation is the document.** The composer lists every Observation on that Citation. Each row names a Subject. Empty Save is allowed. Add property can reuse an existing Citation. Entry points only pre-select.
- **Row-level Observation commits.** Save, revert, and delete one row at a time. Nothing is deleted by omission. Citation fields have their own Save. Leaving with unsaved work always asks.
- **Connect is one composer row.** `CreateCitedBridge` writes the bridge + edges atomically. Endpoints are fixed; a wrong one means discard or delete the bridge. Location connections have no term. The server places the bridge at the endpoint midpoint.
- **Bridges are named by their cited sentence.** Prefer each endpoint’s identity Observation (`name` / `event_type` / `toponym`), then working label. New bridges store no label.
- **Vision is image-only.** Auto Transcribe fills citation transcription from the image (or region). No Observation writes. No Vision on PDF pages.
- **PDF is Preview’s pointer.** Drag selects; scroll pans. Find is navigation-only (highlight + page jump, not a locator). Paste from selection fills transcription. Image-only PDFs fail honestly.
- **Catalog metadata dates are text.** `source_metadata` has no `date_value_id`. Saved values delete via `ClearSourceMetadata`. `url` values are shape-checked and open in the browser.
- **List graph counts are their own cache.** Subject and observation counts live on `.sourceGraphProgress`, not on `sourcesList`.
- **Delete is exists + extra gates + inbound-empty, or a named Impact list.** Facets CASCADE. Official `Delete` calls `deleteimpact.Impact`. The UI instances the DeleteImpact recipe. Register: [`core/database/deleteimpact`](../../../../core/database/deleteimpact).
- **Connection facets release with the bridge.** Official `subjects.Delete` drops edge Observations plus the Connect disambiguation row. Extra Add-property rows still block. Endpoints and the Citation stay. Other Observation deletes of an edge stay `edge_locked`.
- **Seeded properties stay locked.** Unused user properties (and unused source types/fields) can erase. Plugin origin is never researcher trash. Type-bindings (`subject_type_fields`) CASCADE and do not block. Terms are resources; a property with term rows is blocked.

## Refused

PDF Artifact thumbnails, `text_quote` locators, Source-page Find, Change type, adopt/import, incomplete-bridge chrome, tray / minimap / filters / undo, Citation→Observation cascade, a Terms page, and attaching the Impact proto to `Error`.

## Leftovers

| Item | Home |
| --- | --- |
| Census-scale graph density | [`docs/dogfood/ux.md`](../../../dogfood/ux.md) |
| Extract structure from transcription | [`docs/dogfood/ux.md`](../../../dogfood/ux.md) |
| `text_quote` locators | [`text-quote-locators.md`](../../../ideas/text-quote-locators.md) |
| PDF Artifact thumbnails | [`artifact-pdf-thumbnails.md`](../../../ideas/artifact-pdf-thumbnails.md) |
| First-class Source provenance date | [`source-provenance-date.md`](../../../ideas/source-provenance-date.md) |
| Source-to-source commentary | [`source-to-source-relationships.md`](../../../ideas/source-to-source-relationships.md) |
| Audio / video Sources | [`audio-video-sources.md`](../../../ideas/audio-video-sources.md) |
