# Evidence thumbnails: file-type fallbacks, Source-type icons, and primary Artifact

**Status:** idea — PR1 (file-type glyphs) and PR2 (source-type icons) done; PR3 not roadmapped. Related client media work: [`image-optimization.md`](archive/image-optimization.md). Domain: [`source-layer-data-model.md`](../source-layer-data-model.md) §§3, 6–8; vocabulary: [`seeded-vocabulary.md`](../seeded-vocabulary.md) §2.1.

## Problem

### Non-image Files look empty

`derivatives.EnsureThumbnail` only rasterizes common **image** MIME types. PDFs, Office docs, video, and other evidence skip cleanly — Artifact rows and Sources list cells fall back to a generic empty `PVThumbnail`. Researchers still need a **recognizable** stand-in: “this is a PDF,” “this is a video,” not a blank square.

### Fileless / physical Artifacts look empty too

An Artifact with no File (physical book on the shelf, cassette in a box, courthouse register only seen on microfilm) has nothing to rasterize. A MIME glyph does not apply. The **Source type** is the best signal for what kind of evidence it is — birth certificate, oral history, parish register — if types can carry a **representational icon**.

### Sources have many Artifacts; which thumb represents the Source?

A Source can own several Artifacts, each with zero or one primary File, each File optionally with a thumbnail derivative. List rows and Source identity chrome need **one** representation for the Source. Today the list is enriched with *some* `thumbnail_rel_path` (first available raster path in practice) — there is no durable, user-controllable choice when multiple thumbs (or only fallbacks) exist.

## Idea

Three cooperating layers of representation. **Source cover / list / identity** preference (PR2):

```text
1. Raster file_derivatives thumbnail (when present)
2. Source-type icon (`source_types.icon_key` — always set once types require it)
3. File-type MIME glyph (error path if icon_key missing/unknown)
4. Generic empty placeholder
```

**Artifact rows with a File** stay raster → MIME → empty. **Fileless Artifact rows** use the parent Source’s type icon.

### 1. File-type fallback thumbnails

When there is a File but no raster thumbnail derivative, render a **typed fallback** in `PVThumbnail` (or a thin wrapper):

| Kind of stand-in | Examples |
| --- | --- |
| Document badge | PDF, DOC/DOCX, RTF, plain text |
| Spreadsheet / slides | XLS, CSV, PPT (if we care) |
| Media | MP4 / MOV / audio waveform-style glyph |
| Generic file | unknown / `application/octet-stream` |

Visual language: archival (not colorful emoji), e.g. document silhouette with a short type label (`PDF`, `MP4`) in mono. Map from `files.media_type` and/or filename extension; one shared map (design-system helper).

**Not** inventing fake JPEG derivatives for PDFs in Go unless we later add real `pdf_page_preview`. MIME fallback is **UI chrome**, not a `file_derivatives` row.

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

**Not on Sources:** creating or editing a **Source** does **not** pick a free-floating icon. The researcher picks a **Source type**; the type’s `icon_key` comes along. The create/edit Source type picker should **show** each type’s icon for scanability. Changing a Source’s type later inherits that type’s icon for type-icon fallbacks / cover.

**Where it shows**

| Surface | Behavior |
| --- | --- |
| **Create / edit Source type** | Icon picker — this is where `icon_key` is chosen. |
| **Create / edit Source** | Type picker shows icons next to type labels; no separate icon field on the Source. |
| **Fileless Artifact** | Use the parent Source’s type icon as the Artifact row thumbnail (physical-only stand-in). |
| **Source cover (optional)** | User may set the Source-level thumbnail to the **type icon** (not only to an Artifact raster/MIME glyph) — useful when all Artifacts are fileless or when the type mark is a better list identity than a random scan. |
| **Vocabulary / type lists** | Show the icon next to the type label for scanability. |

Icons live in the **design system** (asset pack + `PVSymbol`-like or dedicated `SourceTypeIcon` keyed enum). Catalog stores only the **key**; clients resolve artwork. Plugins later could register extra keys under a namespaced prefix if needed — keep v1 closed set.

### 3. Source primary thumbnail (rollup + user choice)

Persist what supplies the Source’s representative thumbnail — broader than “which Artifact’s File”:

```text
Source cover resolution
  ├── primary Artifact with raster thumb
  ├── primary Artifact with MIME glyph only
  ├── explicit “use Source type icon”
  └── auto: first raster ensured → else type icon → else MIME error path → empty
```

**Default behavior**

- When the **first file-bearing Artifact** is created/ingested under a Source that still uses the type icon, **auto-pin that Artifact** as cover (same persisted state as a manual pin). Later files do not steal the cover.
- Until a file-bearing Artifact exists (or after **Revert to default**), cover is the **Source-type icon**.

**User control**

- **Use as thumbnail** on a file-bearing Artifact row (raster or MIME for that Artifact).
- **Revert to default** on the identity cover context menu → type icon even when rasters exist.
- No separate “auto” mode, badge, or revert-to-auto. Badge is only **Cover**.

**Schema**

- `sources.cover_mode` (`artifact` | `type_icon`) + `sources.primary_artifact_id` (nullable FK, `ON DELETE SET NULL`).
- `source_types.icon_key` — **shipped in PR2**.
- List/Get Source returns `cover_mode`, `primary_artifact_id`, and resolved `thumbnail_*` for cells.

## Why these hang together

| Situation | What the researcher sees |
| --- | --- |
| Scanned JPEG Artifact | Real thumb; can be Source cover |
| PDF Artifact | PDF glyph on Artifact row; Source cover uses that Artifact’s MIME when pinned (first file auto-pins) |
| Fileless “parish register” Artifact | Source-type scroll/book icon |
| Multi-Artifact Source | Explicit primary Artifact or type-icon cover; first file is a one-shot accelerator |

Without type icons, physical evidence stays blank. Without MIME glyphs, digital non-images stay blank. Without primary/cover control, multi-Artifact Sources stay arbitrary.

## Suggested delivery (when scheduled)

**Multiple PRs**, not one mega-PR. The three layers ship value independently, touch different surfaces (Swift-only vs schema+FFI vs cover UX), and leave the tree reviewable. A single PR would mix design-system assets, catalog migrations, and Source identity chrome.

Design assets: PR1 shipped `file_*`; PR2 shipped `type_*`.

| PR | Delivers | Leaves the tree… |
| --- | --- | --- |
| **1 — File-type glyphs** | **Done:** nine `file_*` assets + `PVEvidenceIcon` / `PVFileTypeGlyph`; `PVThumbnail` evidence-glyph state; `CachedThumbnail` MIME fallback; ListSources `thumbnail_media_type` + `thumbnail_original_filename` with first-raster-then-file cover scan; Artifact rows, Sources list, identity header wired | Digital non-images (PDF, video, …) show a typed stand-in; empty only for fileless / unknown-missing |
| **2 — Source-type icons** | **Done:** closed `icon_key` on `source_types`; seed defaults; icon picker on create/edit **type**; type icons in create/edit **Source** type picker; fileless Artifact thumbs + type-list marks; cover order raster → type → MIME → empty | Physical / fileless rows and vocabulary are scannable; custom types default to `type_evidence` |
| **3 — Source primary cover** | **Done:** `cover_mode` + `primary_artifact_id`; first file-bearing Artifact auto-pins; “Use as thumbnail” / “Revert to default”; List/Get enriched cover | Multi-Artifact Sources have durable, controllable list identity |

**Do not** fold cover control into PR1/PR2 — resolution rules and open questions below belong in PR3 once both fallbacks exist.

**Do not** invent `file_derivatives` rows for MIME glyphs — UI chrome only until a real `pdf_page_preview` (non-goal).

### PR1 — File-type fallback glyphs (**done**)

Shipped on `feat/file-type-fallback-glyphs`:

1. Closed `file_*` key set + SVG imagesets under `Assets.xcassets/EvidenceIcons/` (from design-system export).
2. `PVFileTypeGlyph` MIME / extension → key map (Swift tests).
3. `PVThumbnail` evidence-glyph content + `PVEvidenceIcon` (ink-band label composited in SwiftUI).
4. Wired: Sources list, Artifact rows / primary-file card, Source identity header (client fallthrough from artifacts when workspace `source` lacks cover fields).
5. ListSources enrichment: prefer first successful raster; else empty path + first file-bearing Artifact’s `media_type` / `original_filename`. No catalog schema migration. Files list still descoped.

### PR2 — Source-type icon vocabulary (**done**)

Shipped on `feat/source-type-icons`:

1. Closed `type_*` key set (~21) and seeded defaults (`birth_certificate` → `type_certificate`).
2. Catalog migration `000014`: `source_types.icon_key` (TEXT NOT NULL DEFAULT `type_evidence`).
3. Proto / FFI: create/update/list Source types carry `icon_key`; closed-set validation.
4. `PVEvidenceIconKey` type family + assets; unknown catalog keys → `type_evidence`.
5. Source **types** create/edit: icon picker; list leading icon; locked plugin types show icon read-only.
6. Source create/edit: type combo shows icons; identity pill optional leading icon.
7. Fileless Artifact thumbnail: parent Source type `icon_key`.
8. `publishCoverToList` / list / identity: raster → type icon → MIME error path → empty.
9. No `Source.thumbnail_icon_key` — resolve via `sourceTypeID` → `CatalogSourceType.iconKey`.

### PR3 — Source primary thumbnail / cover (**done**)

Shipped on `feat/source-cover-primary`:

1. Migration `000015`: `sources.cover_mode` (`type_icon` default) + `primary_artifact_id`; backfill first file-bearing Artifact; trigger clears mode when primary nulls.
2. `SetCover` / `MaybePinFirstFileCover`; first create-with-file or ingest auto-pins once.
3. Proto/FFI: `Source.cover_mode` / `primary_artifact_id`; `SetSourceCover`; List + GetSourceWorkspace enrich `thumbnail_*` from mode.
4. Artifact row: **Cover** badge + **Use as thumbnail**; identity cover context menu **Revert to default**.
5. Settled rules: primary = Artifact id; type-icon pin survives later files; PDF pin shows MIME on Source; actions on row + identity menu; no auto badge/state.

## Non-goals (for the parked idea)

- Full in-app PDF/page raster pipeline (`pdf_page_preview`) — heavier; glyphs first
- Replacing content-addressed derivatives with UI-only icons for images that *do* have thumbs
- Async decode cache — stays in [`image-optimization.md`](image-optimization.md)
- Per-Artifact custom uploaded icons (type-level curated set is enough for v1)
- Utility / chrome icons beyond `file_*` and `type_*` (add, pin, repository marks — separate if needed)

## Open questions

**Settled in PR3**

- Primary = Artifact id + `cover_mode` (`artifact` | `type_icon`).
- Type-icon pin is never stolen by later files; first-file auto-pin only when still `type_icon`.
- PDF primary shows PDF MIME glyph on Source list/identity (not silent fallthrough to type).
- Actions: Artifact row (“Use as thumbnail” / Cover badge) + identity cover context menu (“Revert to default”).
- Clients get resolved `thumbnail_*` plus `cover_mode` / `primary_artifact_id`; type icon still via `sourceTypeID` → `icon_key`.

## Related docs

- [`source-layer-data-model.md`](../source-layer-data-model.md) §§3, 6–8
- [`seeded-vocabulary.md`](../seeded-vocabulary.md) §2.1 (`source_types`)
- [`image-optimization.md`](archive/image-optimization.md)
- [`artifact-file-storage.md`](../artifact-file-storage.md)
- Design assets: Provenencia Design System export (`provenencia_file_*` / `provenencia_type_*`)
