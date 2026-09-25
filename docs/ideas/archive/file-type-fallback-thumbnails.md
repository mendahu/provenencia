# Evidence thumbnails (archived)

**Done.** Domain: [`source-layer-data-model.md`](../../source-layer-data-model.md) §§3, 6–8.

## Decisions

- Image Files get a raster derivative. Non-image Files get a **file-type glyph**, not an empty thumb.
- Fileless Artifacts use the **Source-type icon**, not a MIME glyph.
- A Source’s cover is **user-pinned** (or empty). Do not auto-pin the first File, and do not paint a Source as a generic PDF icon.
- PDF first-page rasters are parked in [`artifact-pdf-thumbnails.md`](../artifact-pdf-thumbnails.md) (descoped from Spike 8).
