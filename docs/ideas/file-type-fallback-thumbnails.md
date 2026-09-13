# Evidence thumbnails: file-type fallbacks, Source-type icons, and primary Artifact

**Status:** idea — PR1–PR3 shipped; cover rules simplified (no first-file auto-pin; Source never shows file-type glyphs). Related client media work: [`image-optimization.md`](archive/image-optimization.md). Domain: [`source-layer-data-model.md`](../source-layer-data-model.md) §§3, 6–8; vocabulary: [`seeded-vocabulary.md`](../seeded-vocabulary.md) §2.1.

## Problem

### Non-image Files look empty

`derivatives.EnsureThumbnail` only rasterizes common **image** MIME types. PDFs, Office docs, video, and other evidence skip cleanly — Artifact rows fall back to a generic empty `PVThumbnail` without a typed stand-in. Researchers still need a **recognizable** Artifact-row cue: “this is a PDF,” “this is a video,” not a blank square.

### Fileless / physical Artifacts look empty too

An Artifact with no File (physical book on the shelf, cassette in a box, courthouse register only seen on microfilm) has nothing to rasterize. A MIME glyph does not apply. The **Source type** is the best signal for what kind of evidence it is — birth certificate, oral history, parish register — if types can carry a **representational icon**.

### Sources have many Artifacts; which thumb represents the Source?

A Source can own several Artifacts, each with zero or one primary File, each File optionally with a thumbnail derivative. List rows and Source identity chrome need **one** representation for the Source — durable and user-controllable, without borrowing Artifact file-type chrome (a Source with a PDF Artifact should not look like a generic PDF icon in the list).

## Idea

Two surfaces, two fallbacks:

```text
Artifact row (has File):  raster → file-type glyph → empty
Artifact row (fileless):  parent Source type icon → empty

Source cover / list / identity:
  type icon (default)
  XOR explicit pinned Artifact raster
```

Source cover never paints a file-type glyph. Pinning requires a rasterizable Artifact File.

### 1. File-type fallback thumbnails

When there is a File but no raster thumbnail derivative, render a **typed fallback** on **Artifact** surfaces in `PVThumbnail` (or a thin wrapper):

| Kind of stand-in | Examples |
| --- | --- |
| Document badge | PDF, DOC/DOCX, RTF, plain text |
| Spreadsheet / slides | XLS, CSV, PPT (if we care) |
| Media | MP4 / MOV / audio waveform-style glyph |
| Generic file | unknown / `application/octet-stream` |

Visual language: archival (not colorful emoji), e.g. document silhouette with a short type label (`PDF`, `MP4`) in mono. Map from `files.media_type` and/or filename extension; one shared map (design-system helper).

**Not** inventing fake JPEG derivatives for PDFs in Go unless we later add real `pdf_page_preview`. MIME fallback is **UI chrome**, not a `file_derivatives` row — and **not** Source list identity.

### 2. Source-type icon vocabulary

Design a **curated collection** of icons that represent *kinds of evidence*, independent of digital file format. Examples to explore in Design:

- Generic book / bound volume
- Loose document / certificate
- Parchment / scroll
- Cassette / magnetic tape
- Microfilm / microfiche
- Video / film reel
- Audio / interview / oral history
- Photograph / album (even when no File yet)
- Map / plat
- Newspaper / clipping
- Digital-native / website (if useful)
- Generic “evidence” fallback for custom types

**On Source types:** when creating or editing a **type** (seeded or user), the researcher **picks one icon** from the collection (stored as a stable key on `source_types`, e.g. `icon_key`, not raw image bytes). Seeded `provenencia` types ship with sensible defaults; user types default to a generic document/book until changed.

**Not on Sources:** creating or editing a **Source** does **not** pick a free-floating icon. The researcher picks a **Source type**; the type’s `icon_key` comes along. The create/edit Source type picker should **show** each type’s icon for scanability. Changing a Source’s type later inherits that type’s icon for type-icon cover.

**Where it shows**

| Surface | Behavior |
| --- | --- |
| **Create / edit Source type** | Icon picker — this is where `icon_key` is chosen. |
| **Create / edit Source** | Type picker shows icons next to type labels; no separate icon field on the Source. |
| **Fileless Artifact** | Use the parent Source’s type icon as the Artifact row thumbnail (physical-only stand-in). |
| **Source cover (default)** | Source-type icon until the researcher pins a raster Artifact. |
| **Vocabulary / type lists** | Show the icon next to the type label for scanability. |

Icons live in the **design system** (asset pack + `PVEvidenceIcon` keyed enum). Catalog stores only the **key**; clients resolve artwork. Plugins later could register extra keys under a namespaced prefix if needed — keep v1 closed set.

### 3. Source primary thumbnail (explicit raster pin)

Persist what supplies the Source’s representative thumbnail:

```text
Source cover resolution
  ├── type_icon (default) → client paints source_types.icon_key
  └── artifact (explicit) → primary Artifact’s raster JPEG only
```

**Default behavior**

- New Sources stay on the **Source-type icon**.
- Creating or ingesting Files does **not** change cover.
- After **Revert to default**, cover is the type icon again.

**User control**

- **Use as thumbnail** on an Artifact row that already has a **raster** thumbnail (images). PDF / other non-image Files cannot be cover.
- **Revert to default** on the identity cover context menu → type icon.
- No auto-pin, no “auto” mode, badge, or revert-to-auto. Badge is only **Cover**.

**Schema**

- `sources.cover_mode` (`artifact` | `type_icon`) + `sources.primary_artifact_id` (nullable FK, `ON DELETE SET NULL`).
- `source_types.icon_key` — **shipped in PR2**.
- List/Get Source returns `cover_mode`, `primary_artifact_id`, and resolved `thumbnail_rel_path` when the pin has a raster. Source `thumbnail_media_type` / `thumbnail_original_filename` stay unused (wire compat only).

## Why these hang together

| Situation | What the researcher sees |
| --- | --- |
| Scanned JPEG Artifact | Real thumb on the row; **Use as thumbnail** can make it Source cover |
| PDF Artifact | PDF glyph on Artifact row; Source stays type icon (cannot pin) |
| Fileless “parish register” Artifact | Source-type scroll/book icon on the row; Source cover is that same type icon by default |
| Multi-Artifact Source | Explicit raster pin or type-icon cover — never an arbitrary first-file steal |

Without type icons, physical evidence stays blank. Without MIME glyphs, digital non-images stay blank on Artifact rows. Without primary/cover control, multi-Artifact Sources cannot choose which scan represents the Source.

## Suggested delivery (when scheduled)

**Multiple PRs**, not one mega-PR. The three layers ship value independently, touch different surfaces (Swift-only vs schema+FFI vs cover UX), and leave the tree reviewable.

Design assets: PR1 shipped `file_*`; PR2 shipped `type_*`.

| PR | Delivers | Leaves the tree… |
| --- | --- | --- |
| **1 — File-type glyphs** | **Done:** nine `file_*` assets + `PVEvidenceIcon` / `PVFileTypeGlyph`; Artifact rows + MIME map | Digital non-images show a typed stand-in on Artifact rows |
| **2 — Source-type icons** | **Done:** closed `icon_key` on `source_types`; seed defaults; pickers; fileless Artifact thumbs | Physical / fileless rows and vocabulary are scannable |
| **3 — Source primary cover** | **Done (simplified):** `cover_mode` + `primary_artifact_id`; explicit raster-only pin; “Use as thumbnail” / “Revert to default”; no first-file auto-pin; Source never paints MIME | Multi-Artifact Sources have durable, controllable list identity |

**Do not** invent `file_derivatives` rows for MIME glyphs — UI chrome only until a real `pdf_page_preview` (non-goal).

### PR1 — File-type fallback glyphs (**done**)

Shipped on `feat/file-type-fallback-glyphs`:

1. Closed `file_*` key set + SVG imagesets under `Assets.xcassets/EvidenceIcons/`.
2. `PVFileTypeGlyph` MIME / extension → key map (Swift tests).
3. `PVThumbnail` evidence-glyph content + `PVEvidenceIcon`.
4. Wired on Artifact rows / primary-file card. Source list uses type icon / raster only under current cover rules.

### PR2 — Source-type icon vocabulary (**done**)

Shipped on `feat/source-type-icons`:

1. Closed `type_*` key set (~21) and seeded defaults.
2. Catalog migration `000014`: `source_types.icon_key`.
3. Proto / FFI + icon picker on types; type icons in Source type combo; fileless Artifact thumbs.
4. No `Source.thumbnail_icon_key` — resolve via `sourceTypeID` → `CatalogSourceType.iconKey`.

### PR3 — Source primary thumbnail / cover (**done**, then simplified)

Shipped on `feat/source-cover-primary`, then tightened:

1. Migration `000015`: `cover_mode` + `primary_artifact_id` (default `type_icon`); trigger clears mode when primary nulls. No create-time backfill of first file.
2. Migration `000016`: clears any earlier auto-pin / backfill rows to `type_icon`.
3. `SetCover` requires a **raster** thumbnail (not merely file-bearing). `MaybePinFirstFileCover` removed; create/ingest never pin.
4. Proto/FFI: `SetSourceCover`; List + Get enrich `thumbnail_rel_path` from mode when raster exists.
5. Artifact row: **Cover** badge + **Use as thumbnail** (raster only); identity menu **Revert to default**.

## Non-goals (for the parked idea)

- Full in-app PDF/page raster pipeline (`pdf_page_preview`) — heavier; glyphs first
- Replacing content-addressed derivatives with UI-only icons for images that *do* have thumbs
- Async decode cache — stays in [`image-optimization.md`](image-optimization.md)
- Per-Artifact custom uploaded icons (type-level curated set is enough for v1)
- Utility / chrome icons beyond `file_*` and `type_*` (add, pin, repository marks — separate if needed)
- Bubbling Artifact file-type glyphs up to Source list / identity

## Open questions

**Settled**

- Primary = Artifact id + `cover_mode` (`artifact` | `type_icon`).
- No first-file auto-pin; cover changes only via explicit pin / revert / primary delete.
- PDF (and other non-raster) Artifacts cannot be Source cover; they keep MIME glyphs on the Artifact row only.
- Actions: Artifact row (“Use as thumbnail” / Cover badge) + identity cover context menu (“Revert to default”).
- Clients get resolved `thumbnail_rel_path` plus `cover_mode` / `primary_artifact_id`; type icon via `sourceTypeID` → `icon_key`.

## Related docs

- [`source-layer-data-model.md`](../source-layer-data-model.md) §§3, 6–8
- [`seeded-vocabulary.md`](../seeded-vocabulary.md) §2.1 (`source_types`)
- [`image-optimization.md`](archive/image-optimization.md)
- [`artifact-file-storage.md`](../artifact-file-storage.md)
- Design assets: Provenencia Design System export (`provenencia_file_*` / `provenencia_type_*`)
